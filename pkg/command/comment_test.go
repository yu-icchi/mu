package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		msg       string
		expect    Command
		expectErr error
	}{
		{
			name: "plan",
			msg:  "mu plan --project test --workspace dev -- -var=\"test=test\" -var-file=\"value.tfvars\" -destroy",
			expect: &Plan{
				Project:   "test",
				Workspace: "dev",
				Vars: TerraformVars{
					"test=test",
				},
				VarFiles: TerraformVarFiles{
					"value.tfvars",
				},
				Destroy: true,
			},
		},
		{
			name: "apply",
			msg:  "mu apply --project test --workspace dev",
			expect: &Apply{
				Project:   "test",
				Workspace: "dev",
			},
		},
		{
			name: "unlock",
			msg:  "mu unlock --project test --workspace dev --force-unlock LOCK_ID",
			expect: &Unlock{
				Project:       "test",
				Workspace:     "dev",
				ForceUnlockID: "LOCK_ID",
			},
		},
		{
			name:   "help",
			msg:    "mu help",
			expect: &Help{},
		},
		{
			name: "import",
			msg:  "mu import --project test --workspace dev ADDRESS ID -- -var=\"test=test\" -var-file=\"value.tfvars\"",
			expect: &Import{
				Project:   "test",
				Workspace: "dev",
				Address:   "ADDRESS",
				ID:        "ID",
				Vars: TerraformVars{
					"test=test",
				},
				VarFiles: TerraformVarFiles{
					"value.tfvars",
				},
			},
		},
		{
			name: "state rm",
			msg:  "mu state --project test --workspace dev rm ADDRESS -- -dry-run",
			expect: &StateRm{
				Project:   "test",
				Workspace: "dev",
				Addresses: []string{
					"ADDRESS",
				},
				DryRun: true,
			},
		},
		{
			name:      "invalid command",
			msg:       "mu test",
			expectErr: ErrInvalidCommand,
		},
		{
			name:      "invalid command",
			msg:       "message",
			expectErr: ErrInvalidCommand,
		},
		{
			name:      "invalid command",
			msg:       "message\nmessage",
			expectErr: ErrInvalidCommand,
		},
		{
			name:      "invalid command",
			msg:       "test message",
			expectErr: ErrInvalidCommand,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := Parse(tt.msg)
			require.Equal(t, tt.expect, cmd)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}
