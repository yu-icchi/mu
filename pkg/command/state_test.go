package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command   string
		expect    *StateRm
		expectErr error
	}{
		{
			command: "mu state rm ADDRESS",
			expect: &StateRm{
				Addresses: []string{
					"ADDRESS",
				},
			},
		},
		{
			command: "mu state -p test rm ADDRESS",
			expect: &StateRm{
				Project: "test",
				Addresses: []string{
					"ADDRESS",
				},
			},
		},
		{
			command: "mu state -w dev rm ADDRESS",
			expect: &StateRm{
				Workspace: "dev",
				Addresses: []string{
					"ADDRESS",
				},
			},
		},
		{
			command: "mu state --project test --workspace dev rm ADDRESS",
			expect: &StateRm{
				Project:   "test",
				Workspace: "dev",
				Addresses: []string{
					"ADDRESS",
				},
			},
		},
		{
			command: "mu state rm ADDRESS -- -dry-run",
			expect: &StateRm{
				Addresses: []string{
					"ADDRESS",
				},
				DryRun: true,
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
