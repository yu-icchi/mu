package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command   string
		expect    *Plan
		expectErr error
	}{
		{
			command: "mu plan",
			expect:  &Plan{},
		},
		{
			command: "mu plan -p test",
			expect: &Plan{
				Project: "test",
			},
		},
		{
			command: "mu plan -w dev",
			expect: &Plan{
				Workspace: "dev",
			},
		},
		{
			command: "mu plan --project test --workspace dev -- -var 'key=value' -var-file test.tfvars -destroy",
			expect: &Plan{
				Project:   "test",
				Workspace: "dev",
				Vars: TerraformVars{
					"key=value",
				},
				VarFiles: TerraformVarFiles{
					"test.tfvars",
				},
				Destroy: true,
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
