package app

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/yu-icchi/mu/pkg/action"
	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/github"
)

func (a *App) genLockLabel(project string) string {
	const lockLabel = "mu_lock"
	return fmt.Sprintf("%s_%s", lockLabel, project)
}

func (a *App) lock(
	ctx context.Context, project string, prNum int, commandType command.Type, color string,
) error {
	label := a.genLockLabel(project)
	pr, err := a.github.FindPullRequestByLabel(ctx, label)
	if err != nil && !errors.Is(err, github.ErrNotFound) {
		return err
	}
	if pr != nil && pr.Number == prNum {
		return nil
	}
	if pr != nil && pr.Number != prNum {
		lockDesc := fmt.Sprintf("PR: #%d", pr.Number)
		if err := a.notifyLockedMessage(ctx, prNum, commandType, lockDesc); err != nil {
			return err
		}
		return errAlreadyLocked
	}
	lockDesc := fmt.Sprintf("PR: #%d", prNum)
	if err := a.github.CreateLabel(ctx, label, lockDesc, color); err != nil {
		if !github.IsErrAlreadyExists(err) {
			return err
		}
		lockLabel, err := a.github.GetLabel(ctx, label)
		if err != nil {
			return err
		}
		if err := a.notifyLockedMessage(ctx, prNum, commandType, lockLabel.Description); err != nil {
			return err
		}
		return errAlreadyLocked
	}
	return a.github.AddPullRequestLabels(ctx, prNum, []string{label})
}

func (a *App) notifyLockedMessage(ctx context.Context, prNum int, commandType command.Type, description string) error {
	cmdType := cases.Title(language.Und).String(string(commandType))
	lockedMsg := fmt.Sprintf(":lock: **%s Failed** This project is currently locked by %s\nRemove lock label if not needed", cmdType, description)
	return a.github.CreateIssueComment(ctx, prNum, lockedMsg)
}

func (a *App) unlock(ctx context.Context, project string, pr *github.PullRequest) error {
	label := a.genLockLabel(project)
	if !pr.HasLabel(label) {
		return nil
	}
	pullRequests, err := a.github.ListPullRequestsByLabel(ctx, label, 2)
	if err != nil {
		return err
	}
	if len(pullRequests) > 1 {
		if err := a.notifyFailedUnlockMessage(ctx, pr.Number, label); err != nil {
			return err
		}
		return errMultipleLockLabels
	}
	if err := a.github.DeleteLabel(ctx, label); err != nil {
		return err
	}
	unlockedMsg := fmt.Sprintf(":unlock: Unlocked the `%s` project", project)
	return a.github.CreateIssueComment(ctx, pr.Number, unlockedMsg)
}

func (a *App) notifyFailedUnlockMessage(ctx context.Context, prNum int, label string) error {
	url := action.LabelURL(label)
	msg := fmt.Sprintf(":x: **Unlock failed**\nMultiple %s labels exist.\n\n%s", label, url)
	return a.github.CreateIssueComment(ctx, prNum, msg)
}
