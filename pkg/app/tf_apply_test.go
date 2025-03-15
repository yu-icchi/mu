package app

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/terraform"
)

func TestApp_tfApply(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "test/mu")
	t.Setenv("GITHUB_RUN_ID", "test-run-id")
	project := &config.Project{
		Name:      "test",
		Dir:       "./testdata",
		Workspace: "default",
		Terraform: &config.Terraform{
			Version: "1.9.1",
		},
		Plan: &config.Plan{
			Paths: []string{"*.tf*"},
			Auto:  true,
		},
		Apply: &config.Apply{
			RequireApprovals: 0,
		},
	}
	type args struct {
		ctx      context.Context
		prNum    int
		sha      string
		cfg      *config.Project
		artifact *github.Artifact
		reviews  github.Reviews
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    *outputApply
		expectErr error
	}{
		{
			name: "success",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
				m.terraform.EXPECT().Apply(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.ApplyParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.ApplyParams{
							PlanFilePath: "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						return &terraform.Output{
							Result:             "apply result",
							OutsideTerraform:   "",
							ChangedResult:      "",
							Warning:            "",
							HasAddOrUpdateOnly: false,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "apply log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "Apply succeeded.",
					Context:   src,
				}).Return(nil)
			},
			expect: &outputApply{
				result: "apply result",
			},
			expectErr: nil,
		},
		{
			name: "failed terraform init",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
							RawLog:   "init error log",
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
			expectErr: errInitFailed,
		},
		{
			name: "failed terraform apply",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
				m.terraform.EXPECT().Apply(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.ApplyParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.ApplyParams{
							PlanFilePath: "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						return &terraform.Output{
							Result:             "apply error result",
							OutsideTerraform:   "",
							ChangedResult:      "",
							Warning:            "",
							HasAddOrUpdateOnly: false,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           true,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "apply error log",
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
			expectErr: errApplyFailed,
		},
		{
			name: "failed to lock: github.FindPullRequestByLabel",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(nil, assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to update pending status",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
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
			expectErr: assert.AnError,
		},
		{
			name: "failed to download plan file",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						assert.NoError(t, err)
						return assert.AnError
					})
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to decompress archive file",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to terraform setup",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
				m.terraform.EXPECT().Setup(ctx).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to terraform compare version",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
			expectErr: assert.AnError,
		},
		{
			name: "failed to terraform switch workspace",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
			expectErr: assert.AnError,
		},
		{
			name: "failed to terraform init",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
						return nil, assert.AnError
					})
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to terraform apply",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
				m.terraform.EXPECT().Apply(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.ApplyParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.ApplyParams{
							PlanFilePath: "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						return nil, assert.AnError
					})
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "not found plan file",
			args: args{
				ctx:      context.Background(),
				prNum:    1,
				sha:      "test-sha",
				cfg:      project,
				artifact: nil,
				reviews:  github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: errNotFoundPlanFile,
		},
		{
			name: "failed to github.CreateIssueComment",
			args: args{
				ctx:      context.Background(),
				prNum:    1,
				sha:      "test-sha",
				cfg:      project,
				artifact: nil,
				reviews:  github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(assert.AnError)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.FailureStatus,
					TargetURL: actionURL,
					Desc:      "failed.",
					Context:   src,
				}).Return(nil)
			},
			expectErr: assert.AnError,
		},
		{
			name: "failed to update success status",
			args: args{
				ctx:   context.Background(),
				prNum: 1,
				sha:   "test-sha",
				cfg:   project,
				artifact: &github.Artifact{
					ID:   1,
					Name: "mu_test",
				},
				reviews: github.Reviews{},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				actionURL := "https://github.com/test/mu/actions/runs/test-run-id"
				src := "mu/apply: test"
				m.github.EXPECT().FindPullRequestByLabel(ctx, "mu_lock_test").Return(&github.PullRequest{
					ID:             1,
					Number:         1,
					Title:          "title",
					CreatedAt:      time.Time{},
					HeadSHA:        "",
					MergeableState: "clean",
					Labels:         nil,
				}, nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.PendingStatus,
					TargetURL: actionURL,
					Desc:      "in progress...",
					Context:   src,
				}).Return(nil)
				m.github.EXPECT().DownloadArtifact(ctx, int64(1), gomock.Any()).
					DoAndReturn(func(ctx context.Context, artifactID int64, file io.Writer) error {
						_, err := io.Copy(file, strings.NewReader("plan data"))
						return err
					})
				m.archive.EXPECT().Decompress("./testdata", "testdata/test_default_1.tfplan.zip").Return(nil)
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
				m.terraform.EXPECT().Apply(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, params *terraform.ApplyParams, opts ...terraform.Option) (*terraform.Output, error) {
						expectParams := &terraform.ApplyParams{
							PlanFilePath: "test_default_1.tfplan",
						}
						assert.Equal(t, expectParams, params)
						return &terraform.Output{
							Result:             "apply result",
							OutsideTerraform:   "",
							ChangedResult:      "",
							Warning:            "",
							HasAddOrUpdateOnly: false,
							HasDestroy:         false,
							HasNoChanges:       false,
							HasError:           false,
							HasParseError:      false,
							Error:              nil,
							RawLog:             "apply log",
						}, nil
					})
				m.github.EXPECT().ListPullRequestComments(ctx, 1).Return([]*github.Comment{}, nil)
				m.github.EXPECT().CreateIssueComment(ctx, 1, gomock.Any()).Return(nil)
				m.github.EXPECT().CreateCommitStatus(ctx, &github.CommitStatus{
					Sha:       "test-sha",
					Status:    github.SuccessStatus,
					TargetURL: actionURL,
					Desc:      "Apply succeeded.",
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
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			app, mock := newTestAppAndMock(ctrl)
			tt.prepare(tt.args.ctx, mock, t)
			out, err := app.tfApply(tt.args.ctx, tt.args.prNum, tt.args.sha, tt.args.cfg, tt.args.artifact, tt.args.reviews)
			if tt.expectErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expect, out)
		})
	}
}
