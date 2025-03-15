package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command   string
		expect    *Import
		expectErr error
	}{
		{
			command: "mu import ADDRESS ID",
			expect: &Import{
				Address: "ADDRESS",
				ID:      "ID",
			},
		},
		{
			command: "mu import -p test ADDRESS ID",
			expect: &Import{
				Project: "test",
				Address: "ADDRESS",
				ID:      "ID",
			},
		},
		{
			command: "mu import -w dev ADDRESS ID",
			expect: &Import{
				Workspace: "dev",
				Address:   "ADDRESS",
				ID:        "ID",
			},
		},
		{
			command: "mu import --project test --workspace dev ADDRESS ID",
			expect: &Import{
				Project:   "test",
				Workspace: "dev",
				Address:   "ADDRESS",
				ID:        "ID",
			},
		},
		{
			command:   "hoge",
			expectErr: ErrInvalidCommand,
		},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			t.Parallel()
			cmd, err := Parse(tt.command)
			if tt.expectErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.expect, cmd)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}
