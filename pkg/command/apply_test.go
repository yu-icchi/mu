package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command   string
		expect    *Apply
		expectErr error
	}{
		{
			command: "mu apply",
			expect:  &Apply{},
		},
		{
			command: "mu apply -p test",
			expect: &Apply{
				Project: "test",
			},
		},
		{
			command: "mu apply -w dev",
			expect: &Apply{
				Workspace: "dev",
			},
		},
		{
			command: "mu apply --project test --workspace dev",
			expect: &Apply{
				Project:   "test",
				Workspace: "dev",
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
			if tt.expectErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expect, cmd)
			}
		})
	}
}
