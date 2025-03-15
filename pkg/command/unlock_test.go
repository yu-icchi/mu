package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnlock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command   string
		expect    *Unlock
		expectErr error
	}{
		{
			command: "mu unlock",
			expect:  &Unlock{},
		},
		{
			command: "mu unlock -p test",
			expect: &Unlock{
				Project: "test",
			},
		},
		{
			command: "mu unlock -w dev",
			expect: &Unlock{
				Workspace: "dev",
			},
		},
		{
			command: "mu unlock --project test --workspace dev",
			expect: &Unlock{
				Project:   "test",
				Workspace: "dev",
			},
		},
		{
			command: "mu unlock --force-unlock LOCK_ID",
			expect: &Unlock{
				ForceUnlockID: "LOCK_ID",
			},
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
