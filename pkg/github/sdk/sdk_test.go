package sdk

import (
	"testing"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
)

func TestNewActions(t *testing.T) {
	t.Parallel()
	cli := githubv3.NewClient(nil)
	actions := NewActions(cli)
	assert.NotNil(t, actions)
}

func TestNewIssues(t *testing.T) {
	t.Parallel()
	cli := githubv3.NewClient(nil)
	issues := NewIssues(cli)
	assert.NotNil(t, issues)
}

func TestNewPullRequests(t *testing.T) {
	t.Parallel()
	cli := githubv3.NewClient(nil)
	pullRequests := NewPullRequests(cli)
	assert.NotNil(t, pullRequests)
}

func TestNewRepositories(t *testing.T) {
	t.Parallel()
	cli := githubv3.NewClient(nil)
	repositories := NewRepositories(cli)
	assert.NotNil(t, repositories)
}

func TestNewReactions(t *testing.T) {
	t.Parallel()
	cli := githubv3.NewClient(nil)
	reactions := NewReactions(cli)
	assert.NotNil(t, reactions)
}

func TestNewGraphQL(t *testing.T) {
	t.Parallel()
	cli := githubv4.NewClient(nil)
	graphQL := NewGraphQL(cli)
	assert.NotNil(t, graphQL)
}
