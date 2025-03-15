package app

import (
	"context"
	"fmt"

	"github.com/yu-icchi/mu/pkg/action"
	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/terraform"
)

func (a *App) genStatusSource(commandType command.Type, projectName string) string {
	return fmt.Sprintf("mu/%s: %s", commandType, projectName)
}

func (a *App) updatePendingStatus(ctx context.Context, sha, projectName string, commandType command.Type) error {
	const desc = "in progress..."
	url := action.RunURL()
	src := a.genStatusSource(commandType, projectName)
	commitStatus := &github.CommitStatus{
		Sha:       sha,
		Status:    github.PendingStatus,
		TargetURL: url,
		Desc:      desc,
		Context:   src,
	}
	return a.github.CreateCommitStatus(ctx, commitStatus)
}

func (a *App) updateSuccessStatus(ctx context.Context, sha, projectName string, commandType command.Type, output *terraform.Output) error {
	url := action.RunURL()
	src := a.genStatusSource(commandType, projectName)
	var desc string
	switch commandType {
	case command.PlanType:
		desc = output.Result
	case command.ApplyType:
		desc = "Apply succeeded."
	}
	commitStatus := &github.CommitStatus{
		Sha:       sha,
		Status:    github.SuccessStatus,
		TargetURL: url,
		Desc:      desc,
		Context:   src,
	}
	return a.github.CreateCommitStatus(ctx, commitStatus)
}

func (a *App) updateFailureStatus(ctx context.Context, sha, projectName string, commandType command.Type) error {
	const desc = "failed."
	url := action.RunURL()
	src := a.genStatusSource(commandType, projectName)
	commitStatus := &github.CommitStatus{
		Sha:       sha,
		Status:    github.FailureStatus,
		TargetURL: url,
		Desc:      desc,
		Context:   src,
	}
	return a.github.CreateCommitStatus(ctx, commitStatus)
}
