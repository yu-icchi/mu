package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/terraform"
)

func TestApp_tfPlan(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test/mu")
	t.Setenv("GITHUB_RUN_ID", "test-run-id")
	project := &config.Project{
		Name:      "test",
		Dir:       ".",
		Workspace: "default",
		Terraform: &config.Terraform{
			Version: "1.9.1",
		},
		Plan: &config.Plan{
			Paths: []string{"*.tf*"},
			Auto:  true,
		},
	}
	type args struct {
		ctx   context.Context
		prNum int
		sha   string
		cfg   *config.Project
		cmd   *command.Plan
	}
	type expect struct {
		out *outputPlan
		err error
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  expect
	}{
		{
			name: "success auto plan",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog: "init log",
						}, nil
					})
				m.terraform.EXPECT().Plan(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.PlanParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.PlanParams{
							Vars:     nil,
							VarFiles: nil,
							Destroy:  false,
							Out:      "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							Result:             "plan result",
							OutsideTerraform:   "",
							ChangedResult:      "change_result",
							Warning:            "",
							HasAddOrUpdateOnly: true,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "init log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "plan result",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				out: &outputPlan{
					path:   "test_default_1.tfplan",
					result: "plan result",
				},
				err: nil,
			},
		},
		{
			name: "success plan",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd: &command.Plan{
					Project:  "test",
					VarFiles: command.TerraformVarFiles([]string{"value.tfvars"}),
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog: "init log",
						}, nil
					})
				m.terraform.EXPECT().Plan(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.PlanParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.PlanParams{
							Vars: nil,
							VarFiles: []string{
								"value.tfvars",
							},
							Destroy: false,
							Out:     "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							Result:             "plan result",
							OutsideTerraform:   "",
							ChangedResult:      "change_result",
							Warning:            "",
							HasAddOrUpdateOnly: true,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "init log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{
					{
						ID: "test-commit-id-01",
						Author: struct {
							Login string
						}{
							Login: "github-actions",
						},
						Body: "<!-- mu:plan -->\ntest plan log",
					},
					{
						ID: "test-commit-id-02",
						Author: struct {
							Login string
						}{
							Login: "github-actions",
						},
						Body: "message",
					},
					{
						ID: "test-commit-id-03",
						Author: struct {
							Login string
						}{
							Login: "github-actions",
						},
						IsMinimized: true,
						Body:        "message",
					},
					{
						ID: "test-commit-id-04",
						Author: struct {
							Login string
						}{
							Login: "account",
						},
						Body: "message",
					},
				}, nil)
				m.github.EXPECT().HideIssueComment(ctx, "test-commit-id-01").Return(nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "plan result",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				out: &outputPlan{
					path:   "test_default_1.tfplan",
					result: "plan result",
				},
				err: nil,
			},
		},
		{
			name: "success plan failed",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(&terraform.Output{
					RawLog: "init log",
				}, nil)
				m.terraform.EXPECT().Plan(ctx, &terraform.PlanParams{
					Vars:     nil,
					VarFiles: nil,
					Destroy:  false,
					Out:      "test_default_1.tfplan",
				}, gomock.Any()).Return(&terraform.Output{
					Result:             "plan result",
					OutsideTerraform:   "",
					ChangedResult:      "change_result",
					Warning:            "",
					HasAddOrUpdateOnly: true,
					HasDestroy:         false,
					HasNoChanges:       false,
					HasError:           true,
					HasParseError:      false,
					Error:              nil,
					RawLog:             "init log",
				}, nil)
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: errPlanFailed,
			},
		},
		{
			name: "success init failed",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog:   "failed log",
							HasError: true,
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: errInitFailed,
			},
		},
		{
			name: "failed to github.updatePendingStatus",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.FindPullRequestByLabel",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to terraform.Setup",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to terraform.CompareVersion",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to terraform.Init",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to terraform.SwitchWorkspace",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to terraform.Plan",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(&terraform.Output{
					RawLog: "init log",
				}, nil)
				m.terraform.EXPECT().Plan(ctx, &terraform.PlanParams{
					Vars:     nil,
					VarFiles: nil,
					Destroy:  false,
					Out:      "test_default_1.tfplan",
				}, gomock.Any()).Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.ListPullRequestComments",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(&terraform.Output{
					RawLog: "init log",
				}, nil)
				m.terraform.EXPECT().Plan(ctx, &terraform.PlanParams{
					Vars:     nil,
					VarFiles: nil,
					Destroy:  false,
					Out:      "test_default_1.tfplan",
				}, gomock.Any()).Return(&terraform.Output{
					Result:             "result",
					OutsideTerraform:   "",
					ChangedResult:      "change_result",
					Warning:            "",
					HasAddOrUpdateOnly: true,
					HasDestroy:         false,
					HasNoChanges:       false,
					HasError:           false,
					HasParseError:      false,
					Error:              nil,
					RawLog:             "init log",
				}, nil)
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.CreateIssueComment",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(&terraform.Output{
					RawLog: "init log",
				}, nil)
				m.terraform.EXPECT().Plan(ctx, &terraform.PlanParams{
					Vars:     nil,
					VarFiles: nil,
					Destroy:  false,
					Out:      "test_default_1.tfplan",
				}, gomock.Any()).Return(&terraform.Output{
					Result:             "result",
					OutsideTerraform:   "",
					ChangedResult:      "change_result",
					Warning:            "",
					HasAddOrUpdateOnly: true,
					HasDestroy:         false,
					HasNoChanges:       false,
					HasError:           false,
					HasParseError:      false,
					Error:              nil,
					RawLog:             "init log",
				}, nil)
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.CreateCommitStatus succeeded",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, &terraform.InitParams{
					BackendConfig:     nil,
					BackendConfigPath: "",
				}, gomock.Any()).Return(&terraform.Output{
					RawLog: "init log",
				}, nil)
				m.terraform.EXPECT().Plan(ctx, &terraform.PlanParams{
					Vars:     nil,
					VarFiles: nil,
					Destroy:  false,
					Out:      "test_default_1.tfplan",
				}, gomock.Any()).Return(&terraform.Output{
					Result:             "result",
					OutsideTerraform:   "",
					ChangedResult:      "change_result",
					Warning:            "",
					HasAddOrUpdateOnly: true,
					HasDestroy:         false,
					HasNoChanges:       false,
					HasError:           false,
					HasParseError:      false,
					Error:              nil,
					RawLog:             "init log",
				}, nil)
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "result",
					Context:   src,
				}).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.updatePendingStatus panic",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).DoAndReturn(func(ctx context.Context, commitStatus *github.CommitStatus) error {
					panic("panic create commit status")
				})
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: errPanicOccurred,
			},
		},
		{
			name: "failed to update failure status",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   "mu/plan: test",
				}).Return(assert.AnError)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "success auto plan",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog: "init log",
						}, nil
					})
				m.terraform.EXPECT().Plan(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.PlanParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.PlanParams{
							Vars:     nil,
							VarFiles: nil,
							Destroy:  false,
							Out:      "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							Result:             "plan result",
							OutsideTerraform:   "",
							ChangedResult:      "change_result",
							Warning:            "",
							HasAddOrUpdateOnly: true,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "init log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "plan result",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				out: &outputPlan{
					path:   "test_default_1.tfplan",
					result: "plan result",
				},
				err: nil,
			},
		},
		{
			name: "init failed: failed to github.ListPullRequestComments",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog:   "failed log",
							HasError: true,
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "init failed: failed to github.CreateIssueComment",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog:   "failed log",
							HasError: true,
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.HideIssueComment",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd: &command.Plan{
					Project: "test",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog: "init log",
						}, nil
					})
				m.terraform.EXPECT().Plan(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.PlanParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.PlanParams{
							Vars:     nil,
							VarFiles: nil,
							Destroy:  false,
							Out:      "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							Result:             "plan result",
							OutsideTerraform:   "",
							ChangedResult:      "change_result",
							Warning:            "",
							HasAddOrUpdateOnly: true,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "init log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{
					{
						ID: "test-commit-id-01",
						Author: struct {
							Login string
						}{
							Login: "github-actions",
						},
						Body: "<!-- mu:plan -->\ntest plan log",
					},
				}, nil)
				m.github.EXPECT().HideIssueComment(ctx, "test-commit-id-01").Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
		{
			name: "failed to github.CreateIssueComment",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				cmd:   &command.Plan{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/plan: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				m.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "").Return(nil)
				m.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(nil)
				m.terraform.EXPECT().CompareVersion(ctx, "1.9.1").Return(nil)
				m.terraform.EXPECT().SwitchWorkspace(ctx, "default").Return(nil)
				m.terraform.EXPECT().Init(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.InitParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.InitParams{
							BackendConfig:     nil,
							BackendConfigPath: "",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							RawLog: "init log",
						}, nil
					})
				m.terraform.EXPECT().Plan(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.PlanParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.PlanParams{
							Vars:     nil,
							VarFiles: nil,
							Destroy:  false,
							Out:      "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						assert.Len(t, opts, 1)
						return &terraform.Output{
							Result:             "plan result",
							OutsideTerraform:   "",
							ChangedResult:      "change_result",
							Warning:            "",
							HasAddOrUpdateOnly: true,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           true,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "plan failed log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expect: expect{
				err: assert.AnError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, mock := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, mock, t)
			out, err := app.tfPlan(tt.args.ctx, tt.args.prNum, tt.args.sha, tt.args.cfg, tt.args.cmd)
			if tt.expect.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expect.out, out)
		})
	}
}
