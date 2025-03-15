package github

import (
	"encoding/json"
	"os"

	githubv3 "github.com/google/go-github/v69/github"
)

const (
	Opened      = "opened"
	Synchronize = "synchronize"
	Reopened    = "reopened"
	Closed      = "closed"
	Created     = "created"
)

type IssueCommentEvent struct {
	githubv3.IssueCommentEvent
}

func (e *IssueCommentEvent) Number() int {
	return e.GetIssue().GetNumber()
}

type PullRequestEvent struct {
	githubv3.PullRequestEvent
}

func (e *PullRequestEvent) Number() int {
	return e.GetNumber()
}

func (g *github) Event() (Event, error) {
	const (
		githubEventName   = "GITHUB_EVENT_NAME"
		githubEventPath   = "GITHUB_EVENT_PATH"
		eventIssueComment = "issue_comment"
		eventPullRequest  = "pull_request"
	)
	eventName := os.Getenv(githubEventName)
	path := os.Getenv(githubEventPath)

	var event Event
	switch eventName {
	case eventIssueComment:
		event = &IssueCommentEvent{}
	case eventPullRequest:
		event = &PullRequestEvent{}
	}
	if event == nil {
		return nil, errUnsupportedEventType
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	if err := json.NewDecoder(file).Decode(event); err != nil {
		return nil, err
	}
	return event, nil
}
