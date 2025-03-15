package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/moby/patternmatcher"
	"gopkg.in/yaml.v3"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	Version                 int        `yaml:"version" validate:"oneof=1"`
	Projects                []*Project `yaml:"projects" validate:"required,dive,required"`
	defaultTerraformVersion string
}

func (c *Config) GetProject(name string) *Project {
	if c == nil {
		return nil
	}
	for _, project := range c.Projects {
		if project.Name == name {
			return project
		}
	}
	return nil
}

func (c *Config) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	return nil
}

func (c *Config) UnmarshalYAML(unmarshal func(any) error) error {
	type config Config
	if err := unmarshal((*config)(c)); err != nil {
		return err
	}
	for _, project := range c.Projects {
		project.Dir = path.Clean(project.Dir)
		if project.Terraform == nil {
			project.Terraform = &Terraform{}
		}
		if project.Terraform.Version == "" {
			if c.defaultTerraformVersion != "" {
				project.Terraform.Version = c.defaultTerraformVersion
			} else {
				project.Terraform.Version = "latest"
			}
		}
	}
	return nil
}

type Project struct {
	Name           string     `yaml:"name" validate:"required"`
	Dir            string     `yaml:"dir" validate:"required"`
	Workspace      string     `yaml:"workspace"`
	Terraform      *Terraform `yaml:"terraform"`
	Plan           *Plan      `yaml:"plan" validate:"required"`
	Apply          *Apply     `yaml:"apply"`
	LockLabelColor string     `yaml:"lock_label_color"`
}

type Projects []*Project

type Terraform struct {
	Version           string            `yaml:"version"`
	ExecPath          string            `yaml:"exec_path"`
	Vars              []string          `yaml:"vars"`
	VarFiles          []string          `yaml:"var_files"`
	BackendConfigPath string            `yaml:"backend_config_path"`
	BackendConfig     map[string]string `yaml:"backend_config"`
}

func (t *Terraform) GetVersion() string {
	if t == nil {
		return ""
	}
	return t.Version
}

func (t *Terraform) GetExecPath() string {
	if t == nil {
		return ""
	}
	return t.ExecPath
}

func (t *Terraform) GetVars() []string {
	if t == nil {
		return nil
	}
	return t.Vars
}

func (t *Terraform) GetVarFiles() []string {
	if t == nil {
		return nil
	}
	return t.VarFiles
}

func (t *Terraform) GetBackendConfigPath() string {
	if t == nil {
		return ""
	}
	return t.BackendConfigPath
}

func (t *Terraform) GetBackendConfig() map[string]string {
	if t == nil {
		return nil
	}
	return t.BackendConfig
}

type Plan struct {
	Paths []string `yaml:"paths" validate:"required,gt=0,dive,required"`
	Auto  bool     `yaml:"auto"`
}

func (p *Plan) HasMatchedPaths(baseDir string, files []string) bool {
	patterns := make([]string, 0, len(p.Paths))
	for _, path := range p.Paths {
		path = strings.TrimSpace(path)
		var exclusion bool
		if path != "" && path[0] == '!' {
			path = path[1:]
			exclusion = true
		}
		pattern := filepath.Join(baseDir, path)
		if exclusion {
			pattern = "!" + pattern
		}
		patterns = append(patterns, pattern)
	}
	matcher, err := patternmatcher.New(patterns)
	if err != nil {
		return false
	}
	for _, file := range files {
		matched, err := matcher.MatchesOrParentMatches(file)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

type Apply struct {
	RequireApprovals int `yaml:"require_approvals"`
}

func (a *Apply) GetRequireApprovals() int {
	if a == nil {
		return 0
	}
	return a.RequireApprovals
}

type options struct {
	defaultTerraformVersion string
}

type Option func(o *options)

func WithDefaultTerraformVersion(version string) Option {
	return func(o *options) {
		o.defaultTerraformVersion = version
	}
}

func Load(filePath string, opts ...Option) (*Config, error) {
	o := &options{}
	for i := range opts {
		opts[i](o)
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	expanded := os.ExpandEnv(string(file))
	cfg := &Config{
		defaultTerraformVersion: o.defaultTerraformVersion,
	}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
