package artifact

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type uploadArtifactWithOptions struct {
	Name      string `yaml:"name,omitempty"`
	Path      string `yaml:"path,omitempty"`
	Overwrite bool   `yaml:"overwrite,omitempty"`
}

func (u *uploadArtifactWithOptions) withType() string {
	return "upload_artifact"
}

type UploadArtifactParams struct {
	Version   string
	Dir       string
	Artifacts []*Artifact
}

type Artifact struct {
	Name      string
	Path      string
	Overwrite bool
}

func UploadArtifacts(params *UploadArtifactParams) error {
	const (
		using      = "composite"
		uses       = "actions/upload-artifact@"
		actionFile = "action.yaml"
	)
	steps := make([]Setup, 0, len(params.Artifacts))
	for _, artifact := range params.Artifacts {
		step := Setup{
			Name: artifact.Name,
			Uses: uses + params.Version,
			With: &uploadArtifactWithOptions{
				Name:      artifact.Name,
				Path:      artifact.Path,
				Overwrite: artifact.Overwrite,
			},
		}
		steps = append(steps, step)
	}
	action := &Action{
		Runs: Runs{
			Using: using,
			Steps: steps,
		},
	}
	if fd, err := os.Stat(params.Dir); os.IsNotExist(err) || !fd.IsDir() {
		if err := os.Mkdir(params.Dir, 0755); err != nil {
			return err
		}
	}
	file, err := os.Create(filepath.Join(params.Dir, actionFile))
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(action); err != nil {
		return err
	}
	return encoder.Close()
}
