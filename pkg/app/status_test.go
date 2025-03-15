package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/terraform"
)

func TestApp_updatePendingStatus(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test_repo")
	t.Setenv("GITHUB_RUN_ID", "test_run_id")
	type args struct {
		ctx         context.Context
		sha         string
		projectName string
		commandType command.Type
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success: plan",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.PlanType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "in progress...",
					Context:   "mu/plan: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "success: apply",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.ApplyType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "in progress...",
					Context:   "mu/apply: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "failed to create commit status",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.ApplyType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "in progress...",
					Context:   "mu/apply: test-project",
				}).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, m := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, m, t)
			err := app.updatePendingStatus(tt.args.ctx, tt.args.sha, tt.args.projectName, tt.args.commandType)
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApp_updateSuccessStatus(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test_repo")
	t.Setenv("GITHUB_RUN_ID", "test_run_id")
	type args struct {
		ctx         context.Context
		sha         string
		projectName string
		commandType command.Type
		output      *terraform.Output
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success: plan",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.PlanType,
				output: &terraform.Output{
					Result: "Plan: 1 to add, 0 to change, 0 to destroy.",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "Plan: 1 to add, 0 to change, 0 to destroy.",
					Context:   "mu/plan: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "success: plan no change",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.PlanType,
				output: &terraform.Output{
					Result: "No changes. Your infrastructure matches the configuration.",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "No changes. Your infrastructure matches the configuration.",
					Context:   "mu/plan: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "success: apply",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.ApplyType,
				output: &terraform.Output{
					Result: "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "Apply succeeded.",
					Context:   "mu/apply: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "failed to create commit status",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.ApplyType,
				output: &terraform.Output{
					Result: "Success",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "Apply succeeded.",
					Context:   "mu/apply: test-project",
				}).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, m := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, m, t)
			err := app.updateSuccessStatus(tt.args.ctx, tt.args.sha, tt.args.projectName, tt.args.commandType, tt.args.output)
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApp_updateFailureStatus(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test_repo")
	t.Setenv("GITHUB_RUN_ID", "test_run_id")
	type args struct {
		ctx         context.Context
		sha         string
		projectName string
		commandType command.Type
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success: plan",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.PlanType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "failed.",
					Context:   "mu/plan: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "success: apply",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.ApplyType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "failed.",
					Context:   "mu/apply: test-project",
				}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "failed to create commit status",
			args: args{
				ctx:         context.Background(),
				sha:         "test-sha",
				projectName: "test-project",
				commandType: command.PlanType,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: "https://github.com/test_repo/actions/runs/test_run_id",
					Desc:      "failed.",
					Context:   "mu/plan: test-project",
				}).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, m := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, m, t)
			err := app.updateFailureStatus(tt.args.ctx, tt.args.sha, tt.args.projectName, tt.args.commandType)
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}
