package github

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmock "github.com/yu-icchi/mu/pkg/github/sdk/mock"
)

type mock struct {
	client            *http.Client
	actions           *sdkmock.MockActions
	issues            *sdkmock.MockIssues
	pullRequest       *sdkmock.MockPullRequests
	repositories      *sdkmock.MockRepositories
	reactions         *sdkmock.MockReactions
	graphQL           *sdkmock.MockGraphQL
	artifactServerURL string
}

func newMock(ctrl *gomock.Controller) *mock {
	return &mock{
		actions:      sdkmock.NewMockActions(ctrl),
		issues:       sdkmock.NewMockIssues(ctrl),
		pullRequest:  sdkmock.NewMockPullRequests(ctrl),
		repositories: sdkmock.NewMockRepositories(ctrl),
		reactions:    sdkmock.NewMockReactions(ctrl),
		graphQL:      sdkmock.NewMockGraphQL(ctrl),
	}
}

func newTestGithub(mock *mock) Github {
	return &github{
		cli:          mock.client,
		actions:      mock.actions,
		issues:       mock.issues,
		pullRequests: mock.pullRequest,
		repositories: mock.repositories,
		reactions:    mock.reactions,
		graphQL:      mock.graphQL,
		owner:        "test-owner",
		repo:         "test-repo",
	}
}

type prepare func(ctx context.Context, m *mock, t *testing.T)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	token := "token"
	owner := "owner"
	repo := "repo"
	gh, err := New(ctx, token, owner, repo)
	require.NotNil(t, gh)
	require.NoError(t, err)
}

func TestGithub_CreateIssueComment(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
		body   string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				number: 1,
				body:   "message",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().CreateComment(ctx, "test-owner", "test-repo", 1, &githubv3.IssueComment{
					Body: githubv3.Ptr("message"),
				}).Return(&githubv3.IssueComment{}, &githubv3.Response{}, nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				number: 1,
				body:   "message",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().CreateComment(ctx, "test-owner", "test-repo", 1, &githubv3.IssueComment{
					Body: githubv3.Ptr("message"),
				}).Return(nil, nil, assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.CreateIssueComment(ctx, tt.args.number, tt.args.body)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_HideIssueComment(t *testing.T) {
	t.Parallel()
	type args struct {
		nodeID string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				nodeID: "test-node-id",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.graphQL.EXPECT().Mutate(ctx, &struct {
					MinimizeComment struct {
						MinimizedComment struct {
							IsMinimized       githubv4.Boolean
							MinimizedReason   githubv4.String
							ViewerCanMinimize githubv4.Boolean
						}
					} `graphql:"minimizeComment(input:$input)"`
				}{}, githubv4.MinimizeCommentInput{
					Classifier: githubv4.ReportedContentClassifiersOutdated,
					SubjectID:  "test-node-id",
				}, nil).Return(nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				nodeID: "test-node-id",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.graphQL.EXPECT().Mutate(ctx, &struct {
					MinimizeComment struct {
						MinimizedComment struct {
							IsMinimized       githubv4.Boolean
							MinimizedReason   githubv4.String
							ViewerCanMinimize githubv4.Boolean
						}
					} `graphql:"minimizeComment(input:$input)"`
				}{}, githubv4.MinimizeCommentInput{
					Classifier: githubv4.ReportedContentClassifiersOutdated,
					SubjectID:  "test-node-id",
				}, nil).Return(assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.HideIssueComment(ctx, tt.args.nodeID)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_CreateIssueCommentReaction(t *testing.T) {
	t.Parallel()
	type args struct {
		commentID int64
		content   string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				commentID: 1,
				content:   "+1",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.reactions.EXPECT().CreateIssueCommentReaction(ctx, "test-owner", "test-repo", int64(1), "+1").Return(&githubv3.Reaction{}, &githubv3.Response{}, nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				commentID: 1,
				content:   "+1",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.reactions.EXPECT().CreateIssueCommentReaction(ctx, "test-owner", "test-repo", int64(1), "+1").Return(nil, nil, assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.CreateIssueCommentReaction(ctx, tt.args.commentID, tt.args.content)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_CreateLabel(t *testing.T) {
	t.Parallel()
	type args struct {
		name        string
		description string
		color       string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				name:        "test-name",
				description: "test-description",
				color:       "test-color",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().CreateLabel(ctx, "test-owner", "test-repo", &githubv3.Label{
					Name:        githubv3.Ptr("test-name"),
					Description: githubv3.Ptr("test-description"),
					Color:       githubv3.Ptr("test-color"),
				}).Return(&githubv3.Label{}, &githubv3.Response{}, nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				name:        "test-name",
				description: "test-description",
				color:       "test-color",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().CreateLabel(ctx, "test-owner", "test-repo", &githubv3.Label{
					Name:        githubv3.Ptr("test-name"),
					Description: githubv3.Ptr("test-description"),
					Color:       githubv3.Ptr("test-color"),
				}).Return(nil, nil, assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.CreateLabel(ctx, tt.args.name, tt.args.description, tt.args.color)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_DeleteLabel(t *testing.T) {
	t.Parallel()
	type args struct {
		label string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().DeleteLabel(ctx, "test-owner", "test-repo", "test-label").Return(&githubv3.Response{}, nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().DeleteLabel(ctx, "test-owner", "test-repo", "test-label").Return(nil, assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.DeleteLabel(ctx, tt.args.label)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_GetLabel(t *testing.T) {
	t.Parallel()
	type args struct {
		label string
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    *Label
		expectErr error
	}{
		{
			name: "success",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().GetLabel(ctx, "test-owner", "test-repo", "test-label").Return(&githubv3.Label{
					Name:        githubv3.Ptr("test-name"),
					Description: githubv3.Ptr("test-description"),
				}, &githubv3.Response{}, nil)
			},
			expect: &Label{
				Name:        "test-name",
				Description: "test-description",
			},
			expectErr: nil,
		},
		{
			name: "failure",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().GetLabel(ctx, "test-owner", "test-repo", "test-label").Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			label, err := gh.GetLabel(ctx, tt.args.label)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, label)
		})
	}
}

func TestGithub_ListReviews(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    Reviews
		expectErr error
	}{
		{
			name: "success",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().ListReviews(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    0,
					PerPage: 100,
				}).Return([]*githubv3.PullRequestReview{
					{
						User: &githubv3.User{
							Login: githubv3.Ptr("user"),
						},
						State: githubv3.Ptr("approve"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: Reviews{
				{
					UserLogin: "user",
					State:     "approve",
				},
			},
			expectErr: nil,
		},
		{
			name: "success: next page",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().ListReviews(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    0,
					PerPage: 100,
				}).Return([]*githubv3.PullRequestReview{
					{
						User: &githubv3.User{
							Login: githubv3.Ptr("user"),
						},
						State: githubv3.Ptr("approve"),
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.pullRequest.EXPECT().ListReviews(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    100,
					PerPage: 100,
				}).Return([]*githubv3.PullRequestReview{
					{
						User: &githubv3.User{
							Login: githubv3.Ptr("user"),
						},
						State: githubv3.Ptr("approve"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: Reviews{
				{
					UserLogin: "user",
					State:     "approve",
				},
				{
					UserLogin: "user",
					State:     "approve",
				},
			},
			expectErr: nil,
		},
		{
			name: "failure",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().ListReviews(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    0,
					PerPage: 100,
				}).Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			reviews, err := gh.ListReviews(ctx, tt.args.number)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, reviews)
		})
	}
}

func TestGithub_ListPullRequestComments(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    []*Comment
		expectErr error
	}{
		{
			name: "success",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				var query struct {
					Repository struct {
						PullRequest struct {
							Comments struct {
								Nodes    []*Comment
								PageInfo struct {
									EndCursor   githubv4.String
									HasNextPage bool
								}
							} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
						} `graphql:"pullRequest(number: $issueNumber)"`
					} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
				}
				variables := map[string]any{
					"repositoryOwner": githubv4.String("test-owner"),
					"repositoryName":  githubv4.String("test-repo"),
					"issueNumber":     githubv4.Int(1),
					"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
				}
				m.graphQL.EXPECT().Query(ctx, &query, variables).Return(nil)
			},
			expect:    nil,
			expectErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			reviews, err := gh.ListPullRequestComments(ctx, tt.args.number)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, reviews)
		})
	}
}

func TestGithub_ListPullRequestsByLabel(t *testing.T) {
	t.Parallel()
	type args struct {
		label string
		limit int
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    []*PullRequest
		expectErr error
	}{
		{
			name: "success: mismatch label",
			args: args{
				label: "test-label",
				limit: 100,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("test-label-A"),
								Description: githubv3.Ptr("test-description"),
							},
						},
						ID:     githubv3.Ptr(int64(100)),
						Number: githubv3.Ptr(100),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect:    nil,
			expectErr: nil,
		},
		{
			name: "success: next page",
			args: args{
				label: "test-label",
				limit: 2,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("test-label"),
								Description: githubv3.Ptr("test-description"),
							},
						},
						ID:     githubv3.Ptr(int64(100)),
						Number: githubv3.Ptr(100),
						Title:  githubv3.Ptr("test-title-A"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    100,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("test-label"),
								Description: githubv3.Ptr("test-description"),
							},
						},
						ID:     githubv3.Ptr(int64(101)),
						Number: githubv3.Ptr(101),
						Title:  githubv3.Ptr("test-title-B"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: []*PullRequest{
				{
					ID:             100,
					Number:         100,
					Title:          "test-title-A",
					CreatedAt:      time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
					HeadSHA:        "test-head-sha",
					MergeableState: "clean",
					Labels: []*Label{
						{
							Name:        "test-label",
							Description: "test-description",
						},
					},
				},
				{
					ID:             101,
					Number:         101,
					Title:          "test-title-B",
					CreatedAt:      time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
					HeadSHA:        "test-head-sha",
					MergeableState: "clean",
					Labels: []*Label{
						{
							Name:        "test-label",
							Description: "test-description",
						},
					},
				},
			},
		},
		{
			name: "failure",
			args: args{
				label: "test-label",
				limit: 100,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			pullRequests, err := gh.ListPullRequestsByLabel(ctx, tt.args.label, tt.args.limit)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, pullRequests)
		})
	}
}

func TestGithub_FindPullRequestByLabel(t *testing.T) {
	t.Parallel()
	type args struct {
		label string
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    *PullRequest
		expectErr error
	}{
		{
			name: "success",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("label"),
								Description: githubv3.Ptr("description"),
							},
						},
						ID:     githubv3.Ptr(int64(10)),
						Number: githubv3.Ptr(10),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("test-label"),
								Description: githubv3.Ptr("test-description"),
							},
						},
						ID:     githubv3.Ptr(int64(100)),
						Number: githubv3.Ptr(100),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: &PullRequest{
				ID:             100,
				Number:         100,
				Title:          "test-title",
				CreatedAt:      time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
				HeadSHA:        "test-head-sha",
				MergeableState: "clean",
				Labels: []*Label{
					{
						Name:        "test-label",
						Description: "test-description",
					},
				},
			},
			expectErr: nil,
		},
		{
			name: "success: next page",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("label"),
								Description: githubv3.Ptr("description"),
							},
						},
						ID:     githubv3.Ptr(int64(10)),
						Number: githubv3.Ptr(10),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    100,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("test-label"),
								Description: githubv3.Ptr("test-description"),
							},
						},
						ID:     githubv3.Ptr(int64(100)),
						Number: githubv3.Ptr(100),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: &PullRequest{
				ID:             100,
				Number:         100,
				Title:          "test-title",
				CreatedAt:      time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
				HeadSHA:        "test-head-sha",
				MergeableState: "clean",
				Labels: []*Label{
					{
						Name:        "test-label",
						Description: "test-description",
					},
				},
			},
			expectErr: nil,
		},
		{
			name: "not found",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return([]*githubv3.PullRequest{
					{
						Labels: []*githubv3.Label{
							{
								Name:        githubv3.Ptr("label"),
								Description: githubv3.Ptr("description"),
							},
						},
						ID:     githubv3.Ptr(int64(10)),
						Number: githubv3.Ptr(10),
						Title:  githubv3.Ptr("test-title"),
						CreatedAt: &githubv3.Timestamp{
							Time: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
						},
						Head: &githubv3.PullRequestBranch{
							SHA: githubv3.Ptr("test-head-sha"),
						},
						MergeableState: githubv3.Ptr("clean"),
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect:    nil,
			expectErr: ErrNotFound,
		},
		{
			name: "failure",
			args: args{
				label: "test-label",
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().List(ctx, "test-owner", "test-repo", &githubv3.PullRequestListOptions{
					State:     "open",
					Sort:      "created",
					Direction: "asc",
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			pullRequest, err := gh.FindPullRequestByLabel(ctx, tt.args.label)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, pullRequest)
		})
	}
}

func TestGithub_AddPullRequestLabels(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
		labels []string
	}
	tests := []struct {
		name    string
		args    args
		prepare prepare
		expect  error
	}{
		{
			name: "success",
			args: args{
				number: 1,
				labels: []string{
					"test-label",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().AddLabelsToIssue(ctx, "test-owner", "test-repo", 1, []string{"test-label"}).Return([]*githubv3.Label{}, &githubv3.Response{}, nil)
			},
			expect: nil,
		},
		{
			name: "failure",
			args: args{
				number: 1,
				labels: []string{
					"test-label",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.issues.EXPECT().AddLabelsToIssue(ctx, "test-owner", "test-repo", 1, []string{"test-label"}).Return(nil, nil, assert.AnError)
			},
			expect: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.AddPullRequestLabels(ctx, tt.args.number, tt.args.labels)
			require.ErrorIs(t, err, tt.expect)
		})
	}
}

func TestGithub_ListFiles(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    []string
		expectErr error
	}{
		{
			name: "success",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().ListFiles(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    0,
					PerPage: 100,
				}).Return([]*githubv3.CommitFile{
					{
						Status:           githubv3.Ptr("created"),
						Filename:         githubv3.Ptr("test1.tf"),
						PreviousFilename: nil,
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.pullRequest.EXPECT().ListFiles(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    100,
					PerPage: 100,
				}).Return([]*githubv3.CommitFile{
					{
						Status:           githubv3.Ptr("renamed"),
						Filename:         githubv3.Ptr("test2.tf"),
						PreviousFilename: githubv3.Ptr("test.tf"),
					},
				}, &githubv3.Response{}, nil)
			},
			expect:    []string{"test1.tf", "test2.tf", "test.tf"},
			expectErr: nil,
		},
		{
			name: "failure: 404",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().ListFiles(ctx, "test-owner", "test-repo", 1, &githubv3.ListOptions{
					Page:    0,
					PerPage: 100,
				}).Return(nil, nil, &githubv3.ErrorResponse{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
					},
				}).Times(5)
			},
			expectErr: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			files, err := gh.ListFiles(ctx, tt.args.number)
			require.Equal(t, tt.expect, files)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestGithub_CreateCommitStatus(t *testing.T) {
	t.Parallel()
	type args struct {
		commitStatus *CommitStatus
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expectErr error
	}{
		{
			name: "success",
			args: args{
				commitStatus: &CommitStatus{
					Sha:       "test-sha",
					Status:    PendingStatus,
					TargetURL: "test-target-url",
					Desc:      "test-description",
					Context:   "test-context",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.repositories.EXPECT().CreateStatus(ctx, "test-owner", "test-repo", "test-sha", &githubv3.RepoStatus{
					State:       githubv3.Ptr("pending"),
					TargetURL:   githubv3.Ptr("test-target-url"),
					Description: githubv3.Ptr("test-description"),
					Context:     githubv3.Ptr("test-context"),
				}).Return(&githubv3.RepoStatus{}, &githubv3.Response{}, nil)
			},
			expectErr: nil,
		},
		{
			name: "failure",
			args: args{
				commitStatus: &CommitStatus{
					Sha:       "test-sha",
					Status:    PendingStatus,
					TargetURL: "test-target-url",
					Desc:      "test-description",
					Context:   "test-context",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.repositories.EXPECT().CreateStatus(ctx, "test-owner", "test-repo", "test-sha", &githubv3.RepoStatus{
					State:       githubv3.Ptr("pending"),
					TargetURL:   githubv3.Ptr("test-target-url"),
					Description: githubv3.Ptr("test-description"),
					Context:     githubv3.Ptr("test-context"),
				}).Return(nil, nil, assert.AnError)
			},
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.CreateCommitStatus(ctx, tt.args.commitStatus)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestGithub_GetPullRequest(t *testing.T) {
	t.Parallel()
	type args struct {
		number int
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    *PullRequest
		expectErr error
	}{
		{
			name: "success",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().Get(ctx, "test-owner", "test-repo", 1).Return(&githubv3.PullRequest{
					ID:        githubv3.Ptr(int64(1)),
					Number:    githubv3.Ptr(1),
					Title:     githubv3.Ptr("title"),
					CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
					Head: &githubv3.PullRequestBranch{
						SHA: githubv3.Ptr("sha"),
					},
					MergeableState: githubv3.Ptr("unstable"),
					Labels: []*githubv3.Label{
						{
							Name:        githubv3.Ptr("label"),
							Description: githubv3.Ptr("description"),
						},
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: &PullRequest{
				ID:             1,
				Number:         1,
				Title:          "title",
				CreatedAt:      time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC),
				HeadSHA:        "sha",
				MergeableState: "unstable",
				Labels: []*Label{
					{
						Name:        "label",
						Description: "description",
					},
				},
			},
			expectErr: nil,
		},
		{
			name: "not found",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().Get(ctx, "test-owner", "test-repo", 1).Return(nil, nil, &githubv3.ErrorResponse{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
					},
				}).Times(5)
			},
			expect: nil,
			expectErr: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
		},
		{
			name: "failure",
			args: args{
				number: 1,
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.pullRequest.EXPECT().Get(ctx, "test-owner", "test-repo", 1).Return(nil, nil, assert.AnError)
			},
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			pr, err := gh.GetPullRequest(ctx, tt.args.number)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, pr)
		})
	}
}

func TestGithub_MultiGetArtifactsByNames(t *testing.T) {
	t.Parallel()
	type args struct {
		names []string
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    Artifacts
		expectErr error
	}{
		{
			name: "success",
			args: args{
				names: []string{"test.tfplan"},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(1)),
							Name:      githubv3.Ptr("hoge.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(10)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 12, 10, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(20)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: Artifacts{
				"test.tfplan": &Artifact{
					ID:        20,
					Name:      "test.tfplan",
					CreatedAt: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			expectErr: nil,
		},
		{
			name: "success: next page",
			args: args{
				names: []string{"test.tfplan"},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(1)),
							Name:      githubv3.Ptr("hoge.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(10)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 12, 10, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    100,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(20)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
			},
			expect: Artifacts{
				"test.tfplan": &Artifact{
					ID:        20,
					Name:      "test.tfplan",
					CreatedAt: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			expectErr: nil,
		},
		{
			name: "failure",
			args: args{
				names: []string{"test.tfplan"},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			pr, err := gh.MultiGetArtifactsByNames(ctx, tt.args.names)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, pr)
		})
	}
}

func TestGithub_DownloadArtifact(t *testing.T) {
	t.Parallel()
	type args struct {
		id      int64
		handler func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expect    []byte
		expectErr error
	}{
		{
			name: "success",
			args: args{
				id: 1,
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				u, err := url.Parse(m.artifactServerURL)
				require.NoError(t, err)
				m.actions.EXPECT().DownloadArtifact(ctx, "test-owner", "test-repo", int64(1), 3).Return(&url.URL{
					Scheme: u.Scheme,
					Host:   u.Host,
					Path:   "test",
				}, &githubv3.Response{}, nil)
			},
			expect:    []byte("ok"),
			expectErr: nil,
		},
		{
			name: "failed to actions.DownloadArtifact",
			args: args{
				id:      1,
				handler: func(w http.ResponseWriter, r *http.Request) {},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().DownloadArtifact(ctx, "test-owner", "test-repo", int64(1), 3).Return(nil, nil, assert.AnError)
			},
			expect:    nil,
			expectErr: assert.AnError,
		},
		{
			name: "failed to internal server error",
			args: args{
				id: 1,
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("error"))
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				u, err := url.Parse(m.artifactServerURL)
				require.NoError(t, err)
				m.actions.EXPECT().DownloadArtifact(ctx, "test-owner", "test-repo", int64(1), 3).Return(&url.URL{
					Scheme: u.Scheme,
					Host:   u.Host,
					Path:   "test",
				}, &githubv3.Response{}, nil)
			},
			expect:    nil,
			expectErr: errUnexpectedStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ts := httptest.NewServer(http.HandlerFunc(tt.args.handler))
			defer ts.Close()
			m := newMock(ctrl)
			m.client = ts.Client()
			m.artifactServerURL = ts.URL
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			file := new(bytes.Buffer)
			err := gh.DownloadArtifact(ctx, tt.args.id, file)
			require.ErrorIs(t, err, tt.expectErr)
			require.Equal(t, tt.expect, file.Bytes())
		})
	}
}

func TestGithub_DeleteArtifactsByNames(t *testing.T) {
	t.Parallel()
	type args struct {
		names []string
	}
	tests := []struct {
		name      string
		args      args
		prepare   prepare
		expectErr error
	}{
		{
			name: "success",
			args: args{
				names: []string{"test.tfplan"},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(1)),
							Name:      githubv3.Ptr("hoge.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(10)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 12, 10, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(20)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{}, nil)
				m.actions.EXPECT().DeleteArtifact(ctx, "test-owner", "test-repo", int64(10)).Return(&githubv3.Response{}, nil)
				m.actions.EXPECT().DeleteArtifact(ctx, "test-owner", "test-repo", int64(20)).Return(&githubv3.Response{}, nil)
			},
			expectErr: nil,
		},
		{
			name: "success: next page",
			args: args{
				names: []string{"test.tfplan"},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(1)),
							Name:      githubv3.Ptr("hoge.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
						{
							ID:        githubv3.Ptr(int64(10)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 12, 10, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{
					NextPage: 100,
				}, nil)
				m.actions.EXPECT().DeleteArtifact(ctx, "test-owner", "test-repo", int64(10)).Return(&githubv3.Response{}, nil)
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    100,
						PerPage: 100,
					},
				}).Return(&githubv3.ArtifactList{
					Artifacts: []*githubv3.Artifact{
						{
							ID:        githubv3.Ptr(int64(20)),
							Name:      githubv3.Ptr("test.tfplan"),
							CreatedAt: &githubv3.Timestamp{Time: time.Date(2025, 2, 15, 12, 0, 0, 0, time.UTC)},
						},
					},
				}, &githubv3.Response{
					NextPage: 0,
				}, nil)
				m.actions.EXPECT().DeleteArtifact(ctx, "test-owner", "test-repo", int64(20)).Return(&githubv3.Response{}, nil)
			},
			expectErr: nil,
		},
		{
			name: "failed to list artifacts",
			args: args{
				names: []string{
					"test.tfplan",
				},
			},
			prepare: func(ctx context.Context, m *mock, t *testing.T) {
				m.actions.EXPECT().ListArtifacts(ctx, "test-owner", "test-repo", &githubv3.ListArtifactsOptions{
					ListOptions: githubv3.ListOptions{
						Page:    0,
						PerPage: 100,
					},
				}).Return(nil, nil, assert.AnError)
			},
			expectErr: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := newMock(ctrl)
			ctx := context.Background()
			tt.prepare(ctx, m, t)
			gh := newTestGithub(m)
			err := gh.DeleteArtifactsByNames(ctx, tt.args.names)
			require.ErrorIs(t, err, tt.expectErr)
		})
	}
}
