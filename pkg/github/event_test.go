package github

import (
	"testing"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub_Event_IssueCommentEvent(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "issue_comment")
	t.Setenv("GITHUB_EVENT_PATH", "./testdata/event_issue_comment.json")

	gh := &github{}
	event, err := gh.Event()
	require.NoError(t, err)
	issueCommentEvent, ok := event.(*IssueCommentEvent)
	require.True(t, ok)
	expect := &IssueCommentEvent{
		IssueCommentEvent: githubv3.IssueCommentEvent{
			Action: githubv3.Ptr("created"),
			Comment: &githubv3.IssueComment{
				ID:     githubv3.Ptr(int64(1)),
				NodeID: githubv3.Ptr("1"),
				Body:   githubv3.Ptr("mu plan"),
			},
			Issue: &githubv3.Issue{
				ID:     githubv3.Ptr(int64(1)),
				Number: githubv3.Ptr(1),
			},
		},
	}
	assert.Equal(t, expect, issueCommentEvent)
	assert.Equal(t, 1, issueCommentEvent.Number())
}

func TestGithub_Event_PullRequestEvent(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	t.Setenv("GITHUB_EVENT_PATH", "./testdata/event_pull_request.json")

	gh := &github{}
	event, err := gh.Event()
	require.NoError(t, err)
	pullRequestEvent, ok := event.(*PullRequestEvent)
	require.True(t, ok)
	expect := &PullRequestEvent{
		PullRequestEvent: githubv3.PullRequestEvent{
			Action: githubv3.Ptr("opened"),
			Number: githubv3.Ptr(1),
			PullRequest: &githubv3.PullRequest{
				ID:     githubv3.Ptr(int64(1)),
				Number: githubv3.Ptr(1),
			},
		},
	}
	assert.Equal(t, expect, pullRequestEvent)
	assert.Equal(t, 1, pullRequestEvent.Number())
}

func TestGithub_Event_UnknownEvent(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "unknown_event")
	t.Setenv("GITHUB_EVENT_PATH", "./testdata/event_unknown.json")

	gh := &github{}
	event, err := gh.Event()
	require.ErrorIs(t, err, errUnsupportedEventType)
	assert.Nil(t, event)
}
