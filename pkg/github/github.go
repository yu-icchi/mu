package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	githubRatelimit "github.com/gofri/go-github-ratelimit/github_ratelimit"
	githubv3 "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/yu-icchi/mu/pkg/github/sdk"
)

const (
	MaxCommentLen = 65536
	ActionBotName = "github-actions"
)

//go:generate mkdir -p mock
//go:generate mockgen -source=github.go -package=mock -destination=mock/mock.go Github

type Github interface {
	CreateIssueComment(ctx context.Context, number int, body string) error
	HideIssueComment(ctx context.Context, nodeID string) error
	CreateIssueCommentReaction(ctx context.Context, commentID int64, content string) error
	CreateLabel(ctx context.Context, name, description, color string) error
	DeleteLabel(ctx context.Context, label string) error
	GetLabel(ctx context.Context, label string) (*Label, error)
	ListReviews(ctx context.Context, number int) (Reviews, error)
	ListPullRequestComments(ctx context.Context, number int) ([]*Comment, error)
	ListPullRequestsByLabel(ctx context.Context, label string, limit int) ([]*PullRequest, error)
	FindPullRequestByLabel(ctx context.Context, label string) (*PullRequest, error)
	AddPullRequestLabels(ctx context.Context, number int, labels []string) error
	ListFiles(ctx context.Context, number int) ([]string, error)
	CreateCommitStatus(ctx context.Context, commitStatus *CommitStatus) error
	GetPullRequest(ctx context.Context, number int) (*PullRequest, error)
	MultiGetArtifactsByNames(ctx context.Context, names []string) (Artifacts, error)
	DownloadArtifact(ctx context.Context, id int64, file io.Writer) error
	DeleteArtifactsByNames(ctx context.Context, names []string) error
	Event() (Event, error)
}

type Event interface {
	Number() int
}

type PullRequest struct {
	ID             int64
	Number         int
	Title          string
	CreatedAt      time.Time
	HeadSHA        string
	MergeableState string
	Labels         []*Label
}

// IsMergeable
// See: https://github.com/octokit/octokit.net/issues/1763
func (p *PullRequest) IsMergeable() bool {
	switch p.MergeableState {
	case "clean", "unstable", "has_hooks":
		return true
	default:
		return false
	}
}

func (p *PullRequest) HasLabel(labelName string) bool {
	for _, label := range p.Labels {
		if label.Name == labelName {
			return true
		}
	}
	return false
}

type Label struct {
	Name        string
	Description string
}

type IssueComment struct {
	NodeID string
}

type File struct {
	Name   string
	Status string
}

type Review struct {
	UserLogin string
	State     string
}

type Reviews []*Review

func (rs Reviews) Approves() int {
	var num int
	for _, r := range rs {
		if strings.ToLower(r.State) == "approve" {
			num++
		}
	}
	return num
}

type CommitStatus struct {
	Sha       string
	Status    Status
	TargetURL string
	Desc      string
	Context   string
}

type Artifact struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Artifacts map[string]*Artifact

func (a Artifacts) Get(name string) *Artifact {
	artifact, ok := a[name]
	if !ok {
		return nil
	}
	return artifact
}

type Status int

const (
	ErrorStatus = iota
	FailureStatus
	PendingStatus
	SuccessStatus
)

func (s Status) String() string {
	switch s {
	case ErrorStatus:
		return "error"
	case FailureStatus:
		return "failure"
	case PendingStatus:
		return "pending"
	case SuccessStatus:
		return "success"
	default:
		return ""
	}
}

type github struct {
	cli          *http.Client
	actions      sdk.Actions
	issues       sdk.Issues
	pullRequests sdk.PullRequests
	repositories sdk.Repositories
	reactions    sdk.Reactions
	graphQL      sdk.GraphQL
	owner, repo  string
}

func New(ctx context.Context, token, owner, repo string) (Github, error) {
	ratelimitCli, err := githubRatelimit.NewRateLimitWaiterClient(http.DefaultTransport)
	if err != nil {
		return nil, err
	}
	v3 := githubv3.NewClient(ratelimitCli).WithAuthToken(token)
	src := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	cli := oauth2.NewClient(ctx, src)
	v4 := githubv4.NewClient(cli)
	return &github{
		cli:          new(http.Client),
		actions:      sdk.NewActions(v3),
		issues:       sdk.NewIssues(v3),
		pullRequests: sdk.NewPullRequests(v3),
		repositories: sdk.NewRepositories(v3),
		reactions:    sdk.NewReactions(v3),
		graphQL:      sdk.NewGraphQL(v4),
		owner:        owner,
		repo:         repo,
	}, nil
}

func (g *github) CreateIssueComment(ctx context.Context, number int, body string) error {
	comment := &githubv3.IssueComment{
		Body: githubv3.Ptr(body),
	}
	_, _, err := g.issues.CreateComment(ctx, g.owner, g.repo, number, comment)
	return err
}

func (g *github) HideIssueComment(ctx context.Context, nodeID string) error {
	var mutate struct {
		MinimizeComment struct {
			MinimizedComment struct {
				IsMinimized       githubv4.Boolean
				MinimizedReason   githubv4.String
				ViewerCanMinimize githubv4.Boolean
			}
		} `graphql:"minimizeComment(input:$input)"`
	}
	input := githubv4.MinimizeCommentInput{
		Classifier: githubv4.ReportedContentClassifiersOutdated,
		SubjectID:  nodeID,
	}
	return g.graphQL.Mutate(ctx, &mutate, input, nil)
}

func (g *github) CreateIssueCommentReaction(ctx context.Context, commentID int64, content string) error {
	_, _, err := g.reactions.CreateIssueCommentReaction(ctx, g.owner, g.repo, commentID, content)
	return err
}

func (g *github) CreateLabel(ctx context.Context, name, description, color string) error {
	label := &githubv3.Label{
		Name:        githubv3.Ptr(name),
		Description: githubv3.Ptr(description),
	}
	if color != "" {
		label.Color = githubv3.Ptr(color)
	}
	_, _, err := g.issues.CreateLabel(ctx, g.owner, g.repo, label)
	return err
}

func (g *github) DeleteLabel(ctx context.Context, label string) error {
	_, err := g.issues.DeleteLabel(ctx, g.owner, g.repo, label)
	return err
}

func (g *github) GetLabel(ctx context.Context, label string) (*Label, error) {
	ghLabel, _, err := g.issues.GetLabel(ctx, g.owner, g.repo, label)
	if err != nil {
		return nil, err
	}
	return &Label{
		Name:        ghLabel.GetName(),
		Description: ghLabel.GetDescription(),
	}, nil
}

func (g *github) ListReviews(ctx context.Context, number int) (Reviews, error) {
	var page int
	reviews := make(Reviews, 0, 10)
	for {
		pullRequestReviews, nextPage, err := g.listReviews(ctx, number, page)
		if err != nil {
			return nil, err
		}
		for _, pullRequestReview := range pullRequestReviews {
			review := &Review{
				UserLogin: pullRequestReview.GetUser().GetLogin(),
				State:     pullRequestReview.GetState(),
			}
			reviews = append(reviews, review)
		}
		if nextPage == 0 {
			break
		}
		page = nextPage
	}
	return reviews, nil
}

func (g *github) listReviews(ctx context.Context, number, page int) ([]*githubv3.PullRequestReview, int, error) {
	opt := &githubv3.ListOptions{
		Page:    page,
		PerPage: 100,
	}
	pullRequestReviews, resp, err := g.pullRequests.ListReviews(ctx, g.owner, g.repo, number, opt)
	if err != nil {
		return nil, 0, err
	}
	return pullRequestReviews, resp.NextPage, nil
}

type Comment struct {
	ID         string
	DatabaseID int64
	Body       string
	Author     struct {
		Login string
	}
	CreatedAt         string
	IsMinimized       bool
	ViewerCanMinimize bool
}

func (g *github) ListPullRequestComments(ctx context.Context, number int) ([]*Comment, error) {
	var q struct {
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
		"repositoryOwner": githubv4.String(g.owner),
		"repositoryName":  githubv4.String(g.repo),
		"issueNumber":     githubv4.Int(number),    // #nosec G115
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}
	var comments []*Comment
	for {
		if err := g.graphQL.Query(ctx, &q, variables); err != nil {
			return nil, err
		}
		comments = append(comments, q.Repository.PullRequest.Comments.Nodes...)
		if !q.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}
	return comments, nil
}

func (g *github) ListPullRequestsByLabel(ctx context.Context, label string, limit int) ([]*PullRequest, error) {
	var (
		list []*PullRequest
		page int
	)
	for {
		pullRequests, nextPage, err := g.listPullRequests(ctx, page)
		if err != nil {
			return nil, err
		}
		for _, pullRequest := range pullRequests {
			for _, prLabel := range pullRequest.Labels {
				if label != prLabel.GetName() {
					continue
				}
				labels := make([]*Label, len(pullRequest.Labels))
				for i, l := range pullRequest.Labels {
					labels[i] = &Label{
						Name:        l.GetName(),
						Description: l.GetDescription(),
					}
				}
				pr := &PullRequest{
					ID:             pullRequest.GetID(),
					Number:         pullRequest.GetNumber(),
					Title:          pullRequest.GetTitle(),
					CreatedAt:      pullRequest.GetCreatedAt().Time,
					HeadSHA:        pullRequest.GetHead().GetSHA(),
					MergeableState: pullRequest.GetMergeableState(),
					Labels:         labels,
				}
				list = append(list, pr)
				break
			}
			if len(list) >= limit {
				return list, nil
			}
		}
		if nextPage == 0 {
			break
		}
		page = nextPage
	}
	return list, nil
}

func (g *github) FindPullRequestByLabel(ctx context.Context, label string) (*PullRequest, error) {
	var page int
	for {
		pullRequests, nextPage, err := g.listPullRequests(ctx, page)
		if err != nil {
			return nil, err
		}
		for _, pullRequest := range pullRequests {
			for _, prLabel := range pullRequest.Labels {
				if label != prLabel.GetName() {
					continue
				}
				labels := make([]*Label, len(pullRequest.Labels))
				for i, l := range pullRequest.Labels {
					labels[i] = &Label{
						Name:        l.GetName(),
						Description: l.GetDescription(),
					}
				}
				pr := &PullRequest{
					ID:             pullRequest.GetID(),
					Number:         pullRequest.GetNumber(),
					Title:          pullRequest.GetTitle(),
					CreatedAt:      pullRequest.GetCreatedAt().Time,
					HeadSHA:        pullRequest.GetHead().GetSHA(),
					MergeableState: pullRequest.GetMergeableState(),
					Labels:         labels,
				}
				return pr, nil
			}
		}
		if nextPage == 0 {
			break
		}
		page = nextPage
	}
	return nil, ErrNotFound
}

func (g *github) listPullRequests(ctx context.Context, page int) ([]*githubv3.PullRequest, int, error) {
	opt := &githubv3.PullRequestListOptions{
		State:     "open",
		Sort:      "created",
		Direction: "asc",
		ListOptions: githubv3.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	}
	pullRequests, resp, err := g.pullRequests.List(ctx, g.owner, g.repo, opt)
	if err != nil {
		return nil, 0, err
	}
	return pullRequests, resp.NextPage, nil
}

func (g *github) AddPullRequestLabels(ctx context.Context, number int, labels []string) error {
	_, _, err := g.issues.AddLabelsToIssue(ctx, g.owner, g.repo, number, labels)
	return err
}

func (g *github) ListFiles(ctx context.Context, number int) ([]string, error) {
	var (
		page  int
		files []string
	)
	for {
		resp, err := g.listFiles(ctx, number, page)
		if err != nil {
			return nil, err
		}
		for _, commitFile := range resp.commitFiles {
			files = append(files, commitFile.GetFilename())
			if commitFile.GetStatus() == "renamed" {
				files = append(files, commitFile.GetPreviousFilename())
			}
		}
		if resp.nextPage == 0 {
			break
		}
		page = resp.nextPage
	}
	return files, nil
}

type listFileResp struct {
	commitFiles []*githubv3.CommitFile
	nextPage    int
}

func (g *github) listFiles(ctx context.Context, number, page int) (*listFileResp, error) {
	operation := func() (*listFileResp, error) {
		opt := &githubv3.ListOptions{
			Page:    page,
			PerPage: 100,
		}
		commitFiles, resp, err := g.pullRequests.ListFiles(ctx, g.owner, g.repo, number, opt)
		if err != nil {
			// Referenced the implementation of Atlantis.
			if IsErrNotFound(err) {
				return nil, err
			}
			return nil, backoff.Permanent(err)
		}
		return &listFileResp{
			commitFiles: commitFiles,
			nextPage:    resp.NextPage,
		}, nil
	}
	return backoff.Retry(ctx, operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(5),
	)
}

func (g *github) CreateCommitStatus(ctx context.Context, commitStatus *CommitStatus) error {
	repoStatus := &githubv3.RepoStatus{
		State:       githubv3.Ptr(commitStatus.Status.String()),
		TargetURL:   githubv3.Ptr(commitStatus.TargetURL),
		Description: githubv3.Ptr(commitStatus.Desc),
		Context:     githubv3.Ptr(commitStatus.Context),
	}
	_, _, err := g.repositories.CreateStatus(ctx, g.owner, g.repo, commitStatus.Sha, repoStatus)
	return err
}

func (g *github) GetPullRequest(ctx context.Context, number int) (*PullRequest, error) {
	pr, err := g.getPullRequest(ctx, number)
	if err != nil {
		return nil, err
	}
	labels := make([]*Label, len(pr.Labels))
	for i, label := range pr.Labels {
		labels[i] = &Label{
			Name:        label.GetName(),
			Description: label.GetDescription(),
		}
	}
	pullRequest := &PullRequest{
		ID:             pr.GetID(),
		Number:         pr.GetNumber(),
		Title:          pr.GetTitle(),
		CreatedAt:      pr.GetCreatedAt().Time,
		HeadSHA:        pr.GetHead().GetSHA(),
		MergeableState: pr.GetMergeableState(),
		Labels:         labels,
	}
	return pullRequest, nil
}

func (g *github) getPullRequest(ctx context.Context, number int) (*githubv3.PullRequest, error) {
	operation := func() (*githubv3.PullRequest, error) {
		pr, _, err := g.pullRequests.Get(ctx, g.owner, g.repo, number)
		if err != nil {
			// Referenced the implementation of Atlantis.
			if IsErrNotFound(err) {
				return nil, err
			}
			return nil, backoff.Permanent(err)
		}
		return pr, nil
	}
	return backoff.Retry(ctx, operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(5),
	)
}

func (g *github) MultiGetArtifactsByNames(ctx context.Context, names []string) (Artifacts, error) {
	var page int
	ret := make(Artifacts, len(names))
	for {
		artifacts, nextPage, err := g.listArtifacts(ctx, page)
		if err != nil {
			return nil, err
		}
		for _, artifact := range artifacts {
			if !slices.Contains(names, artifact.GetName()) {
				continue
			}
			name := artifact.GetName()
			data, ok := ret[name]
			if !ok {
				ret[name] = &Artifact{
					ID:        artifact.GetID(),
					Name:      artifact.GetName(),
					CreatedAt: artifact.GetCreatedAt().Time,
				}
				continue
			}
			// latest artifact data
			if data.ID < artifact.GetID() {
				ret[name] = &Artifact{
					ID:        artifact.GetID(),
					Name:      artifact.GetName(),
					CreatedAt: artifact.GetCreatedAt().Time,
				}
			}
		}
		if nextPage == 0 {
			break
		}
		page = nextPage
	}
	return ret, nil
}

func (g *github) listArtifacts(ctx context.Context, page int) ([]*githubv3.Artifact, int, error) {
	opts := &githubv3.ListArtifactsOptions{
		ListOptions: githubv3.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	}
	listArtifacts, resp, err := g.actions.ListArtifacts(ctx, g.owner, g.repo, opts)
	if err != nil {
		return nil, 0, err
	}
	return listArtifacts.Artifacts, resp.NextPage, nil
}

func (g *github) DownloadArtifact(ctx context.Context, id int64, file io.Writer) error {
	const maxRedirects = 3
	url, _, err := g.actions.DownloadArtifact(ctx, g.owner, g.repo, id, maxRedirects)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := g.cli.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", errUnexpectedStatus, resp.StatusCode)
	}
	_, err = io.Copy(file, resp.Body)
	return err
}

func (g *github) DeleteArtifactsByNames(ctx context.Context, names []string) error {
	var page int
	for {
		artifacts, nextPage, err := g.listArtifacts(ctx, page)
		if err != nil {
			return err
		}
		for _, artifact := range artifacts {
			if !slices.Contains(names, artifact.GetName()) {
				continue
			}
			if _, err := g.actions.DeleteArtifact(ctx, g.owner, g.repo, artifact.GetID()); err != nil {
				return err
			}
		}
		if nextPage == 0 {
			break
		}
		page = nextPage
	}
	return nil
}
