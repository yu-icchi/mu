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

type outputPlan struct {
	path   string
	result string
}

func (a *App) tfPlan(
	ctx context.Context, prNum int, sha string, projectCfg *config.Project, cmd *command.Plan,
) (out *outputPlan, err error) {
	defer func() {
		rec := recover()
		if err == nil && rec == nil {
			return
		}
		if rec != nil {
			err = fmt.Errorf("%w: %s", errPanicOccurred, rec)
			a.logger.Debug(fmt.Sprintf("plan: %s", err))
		}
		if err := a.updateFailureStatus(ctx, sha, projectCfg.Name, cmd.Type()); err != nil {
			a.logger.Error("failed to update commit state", log.Error(err))
		}
	}()

	if err := a.updatePendingStatus(ctx, sha, projectCfg.Name, cmd.Type()); err != nil {
		return nil, err
	}

	if err := a.lock(ctx, projectCfg.Name, prNum, cmd.Type(), projectCfg.LockLabelColor); err != nil {
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

	a.action.StartGroup(fmt.Sprintf("mu init --project=%s --workspace=%s", projectCfg.Name, projectCfg.Workspace))
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
		if err := a.hidePlanResultComments(ctx, prNum); err != nil {
			return nil, err
		}
		if err := a.outputInitFailedResult(ctx, prNum, projectCfg, initRet); err != nil {
			return nil, err
		}
		return nil, errInitFailed
	}

	filename := a.genPlanFilename(projectCfg.Name, projectCfg.Workspace, prNum)
	varFiles := projectCfg.Terraform.GetVarFiles()
	if len(cmd.VarFiles) > 0 {
		varFiles = append(varFiles, cmd.VarFiles...)
	}
	a.action.StartGroup(fmt.Sprintf("mu plan --project=%s --workspce=%s", projectCfg.Name, projectCfg.Workspace))
	planRet, err := tf.Plan(ctx, &terraform.PlanParams{
		Vars:     append(projectCfg.Terraform.GetVars(), cmd.Vars...),
		VarFiles: varFiles,
		Destroy:  cmd.Destroy,
		Out:      filename,
	}, terraform.WithStream(os.Stdout))
	_, _ = fmt.Fprintln(os.Stdout)
	a.action.EndGroup()
	if err != nil {
		return nil, err
	}
	if !a.disableSummaryLog {
		a.outputPlanSummary(projectCfg, planRet.RawLog)
	}
	if err := a.hidePlanResultComments(ctx, prNum); err != nil {
		return nil, err
	}
	if err := a.outputPlanResult(ctx, prNum, projectCfg, planRet); err != nil {
		return nil, err
	}
	if planRet.HasError {
		return nil, errPlanFailed
	}
	if err := a.updateSuccessStatus(ctx, sha, projectCfg.Name, cmd.Type(), planRet); err != nil {
		return nil, err
	}

	out = &outputPlan{
		path:   filepath.Join(projectCfg.Dir, filename),
		result: planRet.Result,
	}
	return out, nil
}

func (a *App) hidePlanResultComments(ctx context.Context, prNum int) error {
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
		if !strings.HasPrefix(comment.Body, muInitMeta) && !strings.HasPrefix(comment.Body, muPlanMeta) {
			continue
		}
		if err := a.github.HideIssueComment(ctx, comment.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputPlanResult(ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output) error {
	if out.HasError {
		return a.outputPlanFailedResult(ctx, prNum, cfg, out)
	}
	return a.outputPlanSucceededResult(ctx, prNum, cfg, out)
}

func (a *App) outputPlanSucceededResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	comment := a.planSucceededMessage(cfg, out)
	messages := a.splitMessages(strings.NewReader(comment))
	for _, msg := range messages {
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputPlanFailedResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	comment := a.planFailedMessage(cfg, out)
	messages := a.splitMessages(strings.NewReader(comment))
	for _, msg := range messages {
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) outputPlanSummary(cfg *config.Project, log string) {
	summary := new(strings.Builder)
	summary.WriteString("## mu plan\n\n")
	summary.WriteString(fmt.Sprintf("project: `%s` workspace: `%s`\n", cfg.Name, cfg.Workspace))
	summary.WriteString("<details><summary>Show Output</summary>\n")
	summary.WriteString("\n```\n")
	summary.WriteString(log)
	summary.WriteString("\n```\n")
	summary.WriteString("</details>\n")
	_ = a.action.AddStepSummary(summary.String())
}
