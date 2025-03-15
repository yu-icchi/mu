package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/log"
	"github.com/yu-icchi/mu/pkg/terraform"
)

type outputApply struct {
	result string
}

func (a *App) tfApply(
	ctx context.Context, prNum int, sha string,
	projectCfg *config.Project, artifact *github.Artifact, reviews github.Reviews,
) (out *outputApply, err error) {
	defer func() {
		rec := recover()
		if err == nil && rec == nil {
			return
		}
		if rec != nil {
			err = fmt.Errorf("%w: %s", errPanicOccurred, rec)
			a.logger.Debug(fmt.Sprintf("apply: %+v", rec))
		}
		if err := a.updateFailureStatus(ctx, sha, projectCfg.Name, command.ApplyType); err != nil {
			a.logger.Error("failed to update status", log.Error(err))
		}
	}()

	if requireApprovals := projectCfg.Apply.GetRequireApprovals(); requireApprovals > 0 {
		if approvals := reviews.Approves(); requireApprovals > approvals {
			msg := fmt.Sprintf(":x: At least %d approvals are required before running `mu apply`.", requireApprovals)
			if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("not enough approve: require_approvals: %d, count: %d: %w",
				requireApprovals, approvals, errApprovalsRequired)
		}
	}

	if err := a.lock(ctx, projectCfg.Name, prNum, command.ApplyType, projectCfg.LockLabelColor); err != nil {
		return nil, err
	}
	if err := a.updatePendingStatus(ctx, sha, projectCfg.Name, command.ApplyType); err != nil {
		return nil, err
	}

	if artifact == nil {
		const msgTemp = "%s\nThe plan file for the `%s` project is not in the Actions Artifacts. Please run `mu plan` again."
		if err := a.github.CreateIssueComment(ctx, prNum, fmt.Sprintf(msgTemp, muPlanMeta, projectCfg.Name)); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("plan file is not found: %s: %w", projectCfg.Name, errNotFoundPlanFile)
	}

	filename := a.genPlanFilename(projectCfg.Name, projectCfg.Workspace, prNum)
	planFilePath, err := a.downloadPlanFile(ctx, projectCfg.Dir, filename, artifact.ID)
	if err != nil {
		return nil, err
	}
	if err := a.archiver.Decompress(projectCfg.Dir, planFilePath); err != nil {
		return nil, err
	}

	tf := a.genTerraform(projectCfg)
	if err := tf.Setup(ctx); err != nil {
		return nil, err
	}
	if err := tf.CompareVersion(ctx, projectCfg.Terraform.GetVersion()); err != nil {
		return nil, err
	}
	if err := tf.SwitchWorkspace(ctx, projectCfg.Workspace); err != nil {
		return nil, err
	}

	a.action.StartGroup(fmt.Sprintf("mu init --project %s --workspace %s", projectCfg.Name, projectCfg.Workspace))
	initRet, err := tf.Init(ctx, &terraform.InitParams{
		BackendConfig:     projectCfg.Terraform.GetBackendConfig(),
		BackendConfigPath: projectCfg.Terraform.GetBackendConfigPath(),
	}, terraform.WithStream(os.Stdout))
	_, _ = fmt.Fprintln(os.Stdout)
	a.action.EndGroup()
	if err != nil {
		return nil, err
	}
	if initRet.HasError {
		if !a.disableSummaryLog {
			a.outputInitFailedSummary(projectCfg, initRet.RawLog)
		}
		if err := a.hideApplyResultComments(ctx, prNum); err != nil {
			return nil, err
		}
		if err := a.outputInitFailedResult(ctx, prNum, projectCfg, initRet); err != nil {
			return nil, err
		}
		return nil, errInitFailed
	}

	a.action.StartGroup(fmt.Sprintf("mu apply --project %s --workspace %s", projectCfg.Name, projectCfg.Workspace))
	applyRet, err := tf.Apply(ctx, &terraform.ApplyParams{
		PlanFilePath: filename,
	}, terraform.WithStream(os.Stdout))
	_, _ = fmt.Fprintln(os.Stdout)
	a.action.EndGroup()
	if err != nil {
		return nil, err
	}
	if !a.disableSummaryLog {
		a.outputApplySummary(projectCfg, applyRet.RawLog)
	}
	if err := a.hideApplyResultComments(ctx, prNum); err != nil {
		return nil, err
	}
	if err := a.outputApplyResult(ctx, prNum, projectCfg, applyRet); err != nil {
		return nil, err
	}
	if applyRet.HasError {
		return nil, errApplyFailed
	}

	if err := a.updateSuccessStatus(ctx, sha, projectCfg.Name, command.ApplyType, applyRet); err != nil {
		return nil, err
	}
	return &outputApply{
		result: applyRet.Result,
	}, nil
}

func (a *App) downloadPlanFile(ctx context.Context, dir, filename string, artifactID int64) (string, error) {
	path := filepath.Join(dir, filename+".zip")
	artifactFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = artifactFile.Close()
	}()
	if err := a.github.DownloadArtifact(ctx, artifactID, artifactFile); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) hideApplyResultComments(ctx context.Context, prNum int) error {
	comments, err := a.github.ListPullRequestComments(ctx, prNum)
	if err != nil {
		return err
	}
	for _, comment := range comments {
		if comment.Author.Login != github.ActionBotName {
			continue
		}
		if comment.IsMinimized {
			continue
		}
		if !strings.HasPrefix(comment.Body, muInitMeta) && !strings.HasPrefix(comment.Body, muApplyMeta) {
			continue
		}
		if err := a.github.HideIssueComment(ctx, comment.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputApplyResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	if out.HasError {
		return a.outputApplyFailedResult(ctx, prNum, cfg, out)
	}
	return a.outputApplySucceededResult(ctx, prNum, cfg, out)
}

func (a *App) outputApplySucceededResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	comment := a.applySucceededMessage(cfg, out)
	messages := a.splitMessages(strings.NewReader(comment))
	for _, msg := range messages {
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputApplyFailedResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	comment := a.applyFailedMessage(cfg, out)
	messages := a.splitMessages(strings.NewReader(comment))
	for _, msg := range messages {
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputApplySummary(cfg *config.Project, log string) {
	summary := new(strings.Builder)
	summary.WriteString("## mu apply\n\n")
	summary.WriteString(fmt.Sprintf("project: `%s` workspace: `%s`\n", cfg.Name, cfg.Workspace))
	summary.WriteString("<details><summary>Show Output</summary>\n")
	summary.WriteString("\n```\n")
	summary.WriteString(log)
	summary.WriteString("\n```\n")
	summary.WriteString("</details>\n")
	_ = a.action.AddStepSummary(summary.String())
}
