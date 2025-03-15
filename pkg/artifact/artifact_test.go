package artifact

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadArtifact(t *testing.T) {
	t.Parallel()
	params := &UploadArtifactParams{
		Version: "v4",
		Dir:     "./testdata",
		Artifacts: []*Artifact{
			{
				Name:      "test-01",
				Path:      "test-01.tfplan",
				Overwrite: true,
			},
			{
				Name:      "test-02",
				Path:      "test-02.tfplan",
				Overwrite: true,
			},
		},
	}
	err := UploadArtifacts(params)
	require.NoError(t, err)

	file, err := os.ReadFile("./testdata/action.yaml")
	require.NoError(t, err)
	expect := `runs:
  using: composite
  steps:
    - name: test-01
      uses: actions/upload-artifact@v4
      with:
        name: test-01
        path: test-01.tfplan
        overwrite: true
    - name: test-02
      uses: actions/upload-artifact@v4
      with:
        name: test-02
        path: test-02.tfplan
        overwrite: true
`
	require.Equal(t, expect, string(file))
}
