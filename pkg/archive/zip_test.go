package archive

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnzip(t *testing.T) {
	t.Parallel()
	dest, err := os.MkdirTemp("", "mu_unzip")
	require.NoError(t, err)
	src := "./testdata/mu_aws.zip"
	archiver := NewZipArchiver()
	err = archiver.Decompress(dest, src)
	require.NoError(t, err)
	info, err := os.Stat(dest + "/aws.tfplan")
	require.NoError(t, err)
	assert.Equal(t, "aws.tfplan", info.Name())
	assert.False(t, info.IsDir())
}
