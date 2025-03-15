package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		cfg    *Config
		expect error
	}{
		{
			name: "success",
			cfg: &Config{
				Version: 1,
				Projects: []*Project{
					{
						Name:      "test",
						Dir:       ".",
						Workspace: "default",
						Terraform: &Terraform{
							Version: "latest",
						},
						Plan: &Plan{
							Paths: []string{
								"*tf*",
							},
							Auto: true,
						},
						Apply: &Apply{
							RequireApprovals: 1,
						},
					},
				},
			},
		},
		{
			name: "invalid version",
			cfg: &Config{
				Version: 0,
				Projects: []*Project{
					{
						Name:      "test",
						Dir:       ".",
						Workspace: "default",
						Terraform: &Terraform{
							Version: "latest",
						},
						Plan: &Plan{
							Paths: []string{
								"*tf*",
							},
							Auto: true,
						},
						Apply: &Apply{
							RequireApprovals: 1,
						},
					},
				},
			},
			expect: ErrInvalidConfig,
		},
		{
			name: "invalid projects",
			cfg: &Config{
				Version:  0,
				Projects: []*Project{},
			},
			expect: ErrInvalidConfig,
		},
		{
			name: "invalid project.name",
			cfg: &Config{
				Version: 1,
				Projects: []*Project{
					{
						Name:      "",
						Dir:       ".",
						Workspace: "default",
						Terraform: &Terraform{
							Version: "latest",
						},
						Plan: &Plan{
							Paths: []string{
								"*tf*",
							},
							Auto: true,
						},
						Apply: &Apply{
							RequireApprovals: 1,
						},
					},
				},
			},
			expect: ErrInvalidConfig,
		},
		{
			name: "invalid project.dir",
			cfg: &Config{
				Version: 1,
				Projects: []*Project{
					{
						Name:      "test",
						Dir:       "",
						Workspace: "default",
						Terraform: &Terraform{
							Version: "latest",
						},
						Plan: &Plan{
							Paths: []string{
								"*tf*",
							},
							Auto: true,
						},
						Apply: &Apply{
							RequireApprovals: 1,
						},
					},
				},
			},
			expect: ErrInvalidConfig,
		},
		{
			name: "invalid project.plan.paths",
			cfg: &Config{
				Version: 1,
				Projects: []*Project{
					{
						Name:      "test",
						Dir:       ".",
						Workspace: "default",
						Terraform: &Terraform{
							Version: "latest",
						},
						Plan: &Plan{
							Paths: []string{},
							Auto:  true,
						},
						Apply: &Apply{
							RequireApprovals: 1,
						},
					},
				},
			},
			expect: ErrInvalidConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPlan_HasMatchedPaths(t *testing.T) {
	t.Parallel()
	type args struct {
		paths   []string
		baseDir string
		files   []string
	}
	tests := []struct {
		name   string
		args   args
		expect bool
	}{
		{
			name: "success",
			args: args{
				paths: []string{
					"*.tf*",
				},
				baseDir: "",
				files: []string{
					"main.tf",
				},
			},
			expect: true,
		},
		{
			name: "ignore setting",
			args: args{
				paths: []string{
					"!*.tf*",
				},
				baseDir: "",
				files: []string{
					"main.tf",
				},
			},
			expect: false,
		},
		{
			name: "unmatched files",
			args: args{
				paths: []string{
					"*.tf*",
				},
				baseDir: "",
				files: []string{
					"main",
				},
			},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Plan{
				Paths: tt.args.paths,
			}
			flag := cfg.HasMatchedPaths(tt.args.baseDir, tt.args.files)
			assert.Equal(t, tt.expect, flag)
		})
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("TERRAFORM_VERSION", "1.9.3")
	cfg, err := Load("./testdata/mu.yaml", WithDefaultTerraformVersion("1.9.0"))
	require.NoError(t, err)
	expect := &Config{
		defaultTerraformVersion: "1.9.0",
		Version:                 1,
		Projects: Projects{
			{
				Name:      "test",
				Dir:       "test/aws",
				Workspace: "default",
				Terraform: &Terraform{
					Version:  "1.9.3",
					ExecPath: "",
					Vars: []string{
						"key=value",
					},
					VarFiles:          []string{"test.tfvar"},
					BackendConfigPath: "/path/to/backend.conf",
				},
				Plan: &Plan{
					Paths: []string{
						"*.tf*",
					},
					Auto: true,
				},
				Apply: &Apply{
					RequireApprovals: 1,
				},
				LockLabelColor: "",
			},
			{
				Name:      "sample",
				Dir:       "test/sample",
				Workspace: "",
				Terraform: &Terraform{
					Version:  "1.9.0",
					ExecPath: "",
					Vars:     nil,
					VarFiles: nil,
					BackendConfig: map[string]string{
						"bucket": "test-bucket",
						"prefix": "test/state",
					},
				},
				Plan: &Plan{
					Paths: []string{
						"*.tf*",
					},
					Auto: true,
				},
				Apply:          nil,
				LockLabelColor: "",
			},
		},
	}
	assert.Equal(t, expect, cfg)
}
