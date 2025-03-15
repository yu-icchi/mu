package app

import (
	"context"
	"testing"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/github"
)

func TestApp_lock(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx         context.Context
		project     string
		prNum       int
		commandType command.Type
		color       string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(ctx context.Context, mock *mock, t *testing.T)
		expect  error
	}{
		{
			name: "success: lock",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(nil)
				mock.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(nil)
			},
			expect: nil,
		},
		{
			name: "success: locked",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					Number: 1,
				}, nil)
			},
			expect: nil,
		},
		{
			name: "already locked: find pull requests",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					Number: 2,
				}, nil)
				lockedMsg := ":lock: **Plan Failed** This project is currently locked by PR: #2\nRemove lock label if not needed"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, lockedMsg).Return(nil)
			},
			expect: errAlreadyLocked,
		},
		{
			name: "failed to find pull request",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "failed to create label",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "failed to add pull request labels",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(nil)
				mock.github.EXPECT().AddPullRequestLabels(ctx, 1, []string{"mu_lock_test"}).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "already locked: failed to create issue comment",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					Number: 2,
				}, nil)
				lockedMsg := ":lock: **Plan Failed** This project is currently locked by PR: #2\nRemove lock label if not needed"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, lockedMsg).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "already locked: create label",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(&githubv3.ErrorResponse{
					Errors: []githubv3.Error{
						{
							Code: "already_exists",
						},
					},
				})
				mock.github.EXPECT().GetLabel(ctx, "mu_lock_test").Return(&github.Label{
					Name:        "mu_lock_test",
					Description: "PR: #2",
				}, nil)
				lockedMsg := ":lock: **Plan Failed** This project is currently locked by PR: #2\nRemove lock label if not needed"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, lockedMsg).Return(nil)
			},
			expect: errAlreadyLocked,
		},
		{
			name: "already locked: find pull requests",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(&githubv3.ErrorResponse{
					Errors: []githubv3.Error{
						{
							Code: "already_exists",
						},
					},
				})
				mock.github.EXPECT().GetLabel(ctx, "mu_lock_test").Return(nil, assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "already locked: create label: failed to create issue comment",
			args: args{
				ctx:         context.Background(),
				project:     "test",
				prNum:       1,
				commandType: command.PlanType,
				color:       "ff0000",
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, github.ErrNotFound)
				mock.github.EXPECT().CreateLabel(ctx, "mu_lock_test", "PR: #1", "ff0000").Return(&githubv3.ErrorResponse{
					Errors: []githubv3.Error{
						{
							Code: "already_exists",
						},
					},
				})
				mock.github.EXPECT().GetLabel(ctx, "mu_lock_test").Return(&github.Label{
					Name:        "mu_lock_test",
					Description: "PR: #2",
				}, nil)
				lockedMsg := ":lock: **Plan Failed** This project is currently locked by PR: #2\nRemove lock label if not needed"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, lockedMsg).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, mock := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, mock, t)
			err := app.lock(tt.args.ctx, tt.args.project, tt.args.prNum, tt.args.commandType, tt.args.color)
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApp_unlock(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test/test")
	type args struct {
		ctx     context.Context
		project string
		pr      *github.PullRequest
	}
	tests := []struct {
		name    string
		args    args
		prepare func(ctx context.Context, mock *mock, t *testing.T)
		expect  error
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return([]*github.PullRequest{}, nil)
				mock.github.EXPECT().DeleteLabel(ctx, "mu_lock_test").Return(nil)
				unlockedMsg := ":unlock: Unlocked the `test` project"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, unlockedMsg).Return(nil)
			},
			expect: nil,
		},
		{
			name: "has not label",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test2",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {},
			expect:  nil,
		},
		{
			name: "failed to multiple lock labels",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return([]*github.PullRequest{
					{
						ID:     1,
						Number: 1,
						Title:  "test-1",
						Labels: []*github.Label{
							{
								Name: "mu_lock_test",
							},
						},
					},
					{
						ID:     2,
						Number: 2,
						Title:  "test-2",
						Labels: []*github.Label{
							{
								Name: "mu_lock_test",
							},
						},
					},
				}, nil)
				mock.github.EXPECT().CreateIssueComment(ctx, 1, `:x: **Unlock failed**
Multiple mu_lock_test labels exist.

https://github.com/test/test/labels/mu_lock_test`).Return(nil)
			},
			expect: errMultipleLockLabels,
		},
		{
			name: "failed to list pull request by label",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return(nil, assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "failed to notify failed unlock message",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return([]*github.PullRequest{
					{
						ID:     1,
						Number: 1,
						Title:  "test-1",
						Labels: []*github.Label{
							{
								Name: "mu_lock_test",
							},
						},
					},
					{
						ID:     2,
						Number: 2,
						Title:  "test-2",
						Labels: []*github.Label{
							{
								Name: "mu_lock_test",
							},
						},
					},
				}, nil)
				mock.github.EXPECT().CreateIssueComment(ctx, 1, `:x: **Unlock failed**
Multiple mu_lock_test labels exist.

https://github.com/test/test/labels/mu_lock_test`).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "failed to delete label",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return([]*github.PullRequest{}, nil)
				mock.github.EXPECT().DeleteLabel(ctx, "mu_lock_test").Return(assert.AnError)
			},
			expect: assert.AnError,
		},
		{
			name: "failed to create issue comment",
			args: args{
				ctx:     context.Background(),
				project: "test",
				pr: &github.PullRequest{
					Number: 1,
					Labels: []*github.Label{
						{
							Name: "mu_lock_test",
						},
					},
				},
			},
			prepare: func(ctx context.Context, mock *mock, t *testing.T) {
				mock.github.EXPECT().ListPullRequestsByLabel(ctx, "mu_lock_test", 2).Return([]*github.PullRequest{}, nil)
				mock.github.EXPECT().DeleteLabel(ctx, "mu_lock_test").Return(nil)
				unlockedMsg := ":unlock: Unlocked the `test` project"
				mock.github.EXPECT().CreateIssueComment(ctx, 1, unlockedMsg).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, mock := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, mock, t)
			err := app.unlock(tt.args.ctx, tt.args.project, tt.args.pr)
			if tt.expect != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				return
			}
			require.NoError(t, err)
		})
	}
}
