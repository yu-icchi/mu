// Code generated by MockGen. DO NOT EDIT.
// Source: sdk.go
//
// Generated by this command:
//
//	mockgen -source=sdk.go -package=mock -destination=mock/mock.go Actions Issues PullRequests Repositories Reactions GraphQL
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	url "net/url"
	reflect "reflect"

	github "github.com/google/go-github/v69/github"
	githubv4 "github.com/shurcooL/githubv4"
	gomock "go.uber.org/mock/gomock"
)

// MockActions is a mock of Actions interface.
type MockActions struct {
	ctrl     *gomock.Controller
	recorder *MockActionsMockRecorder
	isgomock struct{}
}

// MockActionsMockRecorder is the mock recorder for MockActions.
type MockActionsMockRecorder struct {
	mock *MockActions
}

// NewMockActions creates a new mock instance.
func NewMockActions(ctrl *gomock.Controller) *MockActions {
	mock := &MockActions{ctrl: ctrl}
	mock.recorder = &MockActionsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockActions) EXPECT() *MockActionsMockRecorder {
	return m.recorder
}

// DeleteArtifact mocks base method.
func (m *MockActions) DeleteArtifact(ctx context.Context, owner, repo string, artifactID int64) (*github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteArtifact", ctx, owner, repo, artifactID)
	ret0, _ := ret[0].(*github.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteArtifact indicates an expected call of DeleteArtifact.
func (mr *MockActionsMockRecorder) DeleteArtifact(ctx, owner, repo, artifactID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteArtifact", reflect.TypeOf((*MockActions)(nil).DeleteArtifact), ctx, owner, repo, artifactID)
}

// DownloadArtifact mocks base method.
func (m *MockActions) DownloadArtifact(ctx context.Context, owner, repo string, artifactID int64, maxRedirects int) (*url.URL, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadArtifact", ctx, owner, repo, artifactID, maxRedirects)
	ret0, _ := ret[0].(*url.URL)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// DownloadArtifact indicates an expected call of DownloadArtifact.
func (mr *MockActionsMockRecorder) DownloadArtifact(ctx, owner, repo, artifactID, maxRedirects any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadArtifact", reflect.TypeOf((*MockActions)(nil).DownloadArtifact), ctx, owner, repo, artifactID, maxRedirects)
}

// ListArtifacts mocks base method.
func (m *MockActions) ListArtifacts(ctx context.Context, owner, repo string, opts *github.ListArtifactsOptions) (*github.ArtifactList, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListArtifacts", ctx, owner, repo, opts)
	ret0, _ := ret[0].(*github.ArtifactList)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListArtifacts indicates an expected call of ListArtifacts.
func (mr *MockActionsMockRecorder) ListArtifacts(ctx, owner, repo, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListArtifacts", reflect.TypeOf((*MockActions)(nil).ListArtifacts), ctx, owner, repo, opts)
}

// MockIssues is a mock of Issues interface.
type MockIssues struct {
	ctrl     *gomock.Controller
	recorder *MockIssuesMockRecorder
	isgomock struct{}
}

// MockIssuesMockRecorder is the mock recorder for MockIssues.
type MockIssuesMockRecorder struct {
	mock *MockIssues
}

// NewMockIssues creates a new mock instance.
func NewMockIssues(ctrl *gomock.Controller) *MockIssues {
	mock := &MockIssues{ctrl: ctrl}
	mock.recorder = &MockIssuesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIssues) EXPECT() *MockIssuesMockRecorder {
	return m.recorder
}

// AddLabelsToIssue mocks base method.
func (m *MockIssues) AddLabelsToIssue(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddLabelsToIssue", ctx, owner, repo, number, labels)
	ret0, _ := ret[0].([]*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// AddLabelsToIssue indicates an expected call of AddLabelsToIssue.
func (mr *MockIssuesMockRecorder) AddLabelsToIssue(ctx, owner, repo, number, labels any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLabelsToIssue", reflect.TypeOf((*MockIssues)(nil).AddLabelsToIssue), ctx, owner, repo, number, labels)
}

// CreateComment mocks base method.
func (m *MockIssues) CreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateComment", ctx, owner, repo, number, comment)
	ret0, _ := ret[0].(*github.IssueComment)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateComment indicates an expected call of CreateComment.
func (mr *MockIssuesMockRecorder) CreateComment(ctx, owner, repo, number, comment any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateComment", reflect.TypeOf((*MockIssues)(nil).CreateComment), ctx, owner, repo, number, comment)
}

// CreateLabel mocks base method.
func (m *MockIssues) CreateLabel(ctx context.Context, owner, resp string, label *github.Label) (*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLabel", ctx, owner, resp, label)
	ret0, _ := ret[0].(*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateLabel indicates an expected call of CreateLabel.
func (mr *MockIssuesMockRecorder) CreateLabel(ctx, owner, resp, label any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLabel", reflect.TypeOf((*MockIssues)(nil).CreateLabel), ctx, owner, resp, label)
}

// DeleteLabel mocks base method.
func (m *MockIssues) DeleteLabel(ctx context.Context, owner, repo, name string) (*github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLabel", ctx, owner, repo, name)
	ret0, _ := ret[0].(*github.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteLabel indicates an expected call of DeleteLabel.
func (mr *MockIssuesMockRecorder) DeleteLabel(ctx, owner, repo, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLabel", reflect.TypeOf((*MockIssues)(nil).DeleteLabel), ctx, owner, repo, name)
}

// GetLabel mocks base method.
func (m *MockIssues) GetLabel(ctx context.Context, owner, repo, name string) (*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLabel", ctx, owner, repo, name)
	ret0, _ := ret[0].(*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetLabel indicates an expected call of GetLabel.
func (mr *MockIssuesMockRecorder) GetLabel(ctx, owner, repo, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLabel", reflect.TypeOf((*MockIssues)(nil).GetLabel), ctx, owner, repo, name)
}

// MockPullRequests is a mock of PullRequests interface.
type MockPullRequests struct {
	ctrl     *gomock.Controller
	recorder *MockPullRequestsMockRecorder
	isgomock struct{}
}

// MockPullRequestsMockRecorder is the mock recorder for MockPullRequests.
type MockPullRequestsMockRecorder struct {
	mock *MockPullRequests
}

// NewMockPullRequests creates a new mock instance.
func NewMockPullRequests(ctrl *gomock.Controller) *MockPullRequests {
	mock := &MockPullRequests{ctrl: ctrl}
	mock.recorder = &MockPullRequestsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPullRequests) EXPECT() *MockPullRequestsMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockPullRequests) Get(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, owner, repo, number)
	ret0, _ := ret[0].(*github.PullRequest)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Get indicates an expected call of Get.
func (mr *MockPullRequestsMockRecorder) Get(ctx, owner, repo, number any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPullRequests)(nil).Get), ctx, owner, repo, number)
}

// List mocks base method.
func (m *MockPullRequests) List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, owner, repo, opts)
	ret0, _ := ret[0].([]*github.PullRequest)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *MockPullRequestsMockRecorder) List(ctx, owner, repo, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPullRequests)(nil).List), ctx, owner, repo, opts)
}

// ListFiles mocks base method.
func (m *MockPullRequests) ListFiles(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFiles", ctx, owner, repo, number, opts)
	ret0, _ := ret[0].([]*github.CommitFile)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListFiles indicates an expected call of ListFiles.
func (mr *MockPullRequestsMockRecorder) ListFiles(ctx, owner, repo, number, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFiles", reflect.TypeOf((*MockPullRequests)(nil).ListFiles), ctx, owner, repo, number, opts)
}

// ListReviews mocks base method.
func (m *MockPullRequests) ListReviews(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.PullRequestReview, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListReviews", ctx, owner, repo, number, opts)
	ret0, _ := ret[0].([]*github.PullRequestReview)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListReviews indicates an expected call of ListReviews.
func (mr *MockPullRequestsMockRecorder) ListReviews(ctx, owner, repo, number, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListReviews", reflect.TypeOf((*MockPullRequests)(nil).ListReviews), ctx, owner, repo, number, opts)
}

// MockRepositories is a mock of Repositories interface.
type MockRepositories struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoriesMockRecorder
	isgomock struct{}
}

// MockRepositoriesMockRecorder is the mock recorder for MockRepositories.
type MockRepositoriesMockRecorder struct {
	mock *MockRepositories
}

// NewMockRepositories creates a new mock instance.
func NewMockRepositories(ctrl *gomock.Controller) *MockRepositories {
	mock := &MockRepositories{ctrl: ctrl}
	mock.recorder = &MockRepositoriesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositories) EXPECT() *MockRepositoriesMockRecorder {
	return m.recorder
}

// CreateStatus mocks base method.
func (m *MockRepositories) CreateStatus(ctx context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateStatus", ctx, owner, repo, ref, status)
	ret0, _ := ret[0].(*github.RepoStatus)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateStatus indicates an expected call of CreateStatus.
func (mr *MockRepositoriesMockRecorder) CreateStatus(ctx, owner, repo, ref, status any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateStatus", reflect.TypeOf((*MockRepositories)(nil).CreateStatus), ctx, owner, repo, ref, status)
}

// MockReactions is a mock of Reactions interface.
type MockReactions struct {
	ctrl     *gomock.Controller
	recorder *MockReactionsMockRecorder
	isgomock struct{}
}

// MockReactionsMockRecorder is the mock recorder for MockReactions.
type MockReactionsMockRecorder struct {
	mock *MockReactions
}

// NewMockReactions creates a new mock instance.
func NewMockReactions(ctrl *gomock.Controller) *MockReactions {
	mock := &MockReactions{ctrl: ctrl}
	mock.recorder = &MockReactionsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReactions) EXPECT() *MockReactionsMockRecorder {
	return m.recorder
}

// CreateIssueCommentReaction mocks base method.
func (m *MockReactions) CreateIssueCommentReaction(ctx context.Context, owner, repo string, commentID int64, content string) (*github.Reaction, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateIssueCommentReaction", ctx, owner, repo, commentID, content)
	ret0, _ := ret[0].(*github.Reaction)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateIssueCommentReaction indicates an expected call of CreateIssueCommentReaction.
func (mr *MockReactionsMockRecorder) CreateIssueCommentReaction(ctx, owner, repo, commentID, content any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssueCommentReaction", reflect.TypeOf((*MockReactions)(nil).CreateIssueCommentReaction), ctx, owner, repo, commentID, content)
}

// MockGraphQL is a mock of GraphQL interface.
type MockGraphQL struct {
	ctrl     *gomock.Controller
	recorder *MockGraphQLMockRecorder
	isgomock struct{}
}

// MockGraphQLMockRecorder is the mock recorder for MockGraphQL.
type MockGraphQLMockRecorder struct {
	mock *MockGraphQL
}

// NewMockGraphQL creates a new mock instance.
func NewMockGraphQL(ctrl *gomock.Controller) *MockGraphQL {
	mock := &MockGraphQL{ctrl: ctrl}
	mock.recorder = &MockGraphQLMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGraphQL) EXPECT() *MockGraphQLMockRecorder {
	return m.recorder
}

// Mutate mocks base method.
func (m *MockGraphQL) Mutate(ctx context.Context, mutate any, input githubv4.Input, variables map[string]any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Mutate", ctx, mutate, input, variables)
	ret0, _ := ret[0].(error)
	return ret0
}

// Mutate indicates an expected call of Mutate.
func (mr *MockGraphQLMockRecorder) Mutate(ctx, mutate, input, variables any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Mutate", reflect.TypeOf((*MockGraphQL)(nil).Mutate), ctx, mutate, input, variables)
}

// Query mocks base method.
func (m *MockGraphQL) Query(ctx context.Context, query any, variables map[string]any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", ctx, query, variables)
	ret0, _ := ret[0].(error)
	return ret0
}

// Query indicates an expected call of Query.
func (mr *MockGraphQLMockRecorder) Query(ctx, query, variables any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockGraphQL)(nil).Query), ctx, query, variables)
}
