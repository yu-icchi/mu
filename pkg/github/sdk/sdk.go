package sdk

import (
	"context"
	"net/url"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
)

//go:generate mkdir -p mock
//go:generate mockgen -source=sdk.go -package=mock -destination=mock/mock.go Actions Issues PullRequests Repositories Reactions GraphQL

type Actions interface {
	ListArtifacts(ctx context.Context, owner, repo string, opts *githubv3.ListArtifactsOptions) (*githubv3.ArtifactList, *githubv3.Response, error)
	DownloadArtifact(ctx context.Context, owner, repo string, artifactID int64, maxRedirects int) (*url.URL, *githubv3.Response, error)
	DeleteArtifact(ctx context.Context, owner, repo string, artifactID int64) (*githubv3.Response, error)
}

type Issues interface {
	CreateComment(ctx context.Context, owner, repo string, number int, comment *githubv3.IssueComment) (*githubv3.IssueComment, *githubv3.Response, error)
	CreateLabel(ctx context.Context, owner, resp string, label *githubv3.Label) (*githubv3.Label, *githubv3.Response, error)
	DeleteLabel(ctx context.Context, owner, repo, name string) (*githubv3.Response, error)
	GetLabel(ctx context.Context, owner, repo, name string) (*githubv3.Label, *githubv3.Response, error)
	AddLabelsToIssue(ctx context.Context, owner, repo string, number int, labels []string) ([]*githubv3.Label, *githubv3.Response, error)
}

type PullRequests interface {
	ListReviews(ctx context.Context, owner, repo string, number int, opts *githubv3.ListOptions) ([]*githubv3.PullRequestReview, *githubv3.Response, error)
	List(ctx context.Context, owner, repo string, opts *githubv3.PullRequestListOptions) ([]*githubv3.PullRequest, *githubv3.Response, error)
	ListFiles(ctx context.Context, owner, repo string, number int, opts *githubv3.ListOptions) ([]*githubv3.CommitFile, *githubv3.Response, error)
	Get(ctx context.Context, owner, repo string, number int) (*githubv3.PullRequest, *githubv3.Response, error)
}

type Repositories interface {
	CreateStatus(ctx context.Context, owner, repo, ref string, status *githubv3.RepoStatus) (*githubv3.RepoStatus, *githubv3.Response, error)
}

type Reactions interface {
	CreateIssueCommentReaction(ctx context.Context, owner, repo string, commentID int64, content string) (*githubv3.Reaction, *githubv3.Response, error)
}

type GraphQL interface {
	Query(ctx context.Context, query any, variables map[string]any) error
	Mutate(ctx context.Context, mutate any, input githubv4.Input, variables map[string]any) error
}

func NewActions(cli *githubv3.Client) Actions {
	return cli.Actions
}

func NewIssues(cli *githubv3.Client) Issues {
	return cli.Issues
}

func NewPullRequests(cli *githubv3.Client) PullRequests {
	return cli.PullRequests
}

func NewRepositories(cli *githubv3.Client) Repositories {
	return cli.Repositories
}

func NewReactions(cli *githubv3.Client) Reactions {
	return cli.Reactions
}

func NewGraphQL(cli *githubv4.Client) GraphQL {
	return cli
}
