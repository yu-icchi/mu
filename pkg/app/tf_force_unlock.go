package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/terraform"
)

func (a *App) tfForceUnlock(ctx context.Context, prNum int, cfg *config.Project, cmd *command.Unlock) error {
	tf := a.genTerraform(cfg)
	if err := tf.Setup(ctx); err != nil {
		return err
	}
	if err := tf.CompareVersion(ctx, cfg.Terraform.GetVersion()); err != nil {
		return err
	}
	if err := tf.SwitchWorkspace(ctx, cfg.Workspace); err != nil {
		return err
	}

	a.action.StartGroup(fmt.Sprintf("mu init --project=%s --workspace=%s", cfg.Name, cfg.Workspace))
	initRet, err := tf.Init(ctx, &terraform.InitParams{
		BackendConfig:     cfg.Terraform.GetBackendConfig(),
		BackendConfigPath: cfg.Terraform.GetBackendConfigPath(),
	}, terraform.WithStream(os.Stdout))
	_, _ = fmt.Fprintln(os.Stdout)
	a.action.EndGroup()
	if err != nil {
		return err
	}
	if initRet.HasError {
		if !a.disableSummaryLog {
			a.outputInitFailedSummary(cfg, initRet.RawLog)
		}
		if err := a.outputInitFailedResult(ctx, prNum, cfg, initRet); err != nil {
			return err
		}
		return errInitFailed
	}

	a.action.StartGroup(fmt.Sprintf("mu unlock --force-unlock %s", cmd.ForceUnlockID))
	forceUnlockRet, err := tf.ForceUnlock(ctx, cmd.ForceUnlockID,
		terraform.WithStream(os.Stdout))
	a.action.EndGroup()
	_, _ = fmt.Fprintln(os.Stdout)
	if err != nil {
		return err
	}
	if !a.disableSummaryLog {
		a.outputForceUnlockSummary(cfg, forceUnlockRet.Result)
	}
	message := a.forceUnlockMessage(forceUnlockRet)
	if err := a.github.CreateIssueComment(ctx, prNum, message); err != nil {
		return fmt.Errorf("%w: %w", errForceUnlockFailed, err)
	}
	if forceUnlockRet.HasError {
		return errForceUnlockFailed
	}
	return nil
}

func (a *App) outputForceUnlockSummary(cfg *config.Project, log string) {
	summary := new(strings.Builder)
	summary.WriteString("## mu force unlock\n\n")
	summary.WriteString(fmt.Sprintf("project: `%s` workspace: `%s`\n", cfg.Name, cfg.Workspace))
	summary.WriteString("<details><summary>Show Output</summary>\n")
	summary.WriteString("\n```\n")
	summary.WriteString(log)
	summary.WriteString("\n```\n")
	summary.WriteString("</details>")
	_ = a.action.AddStepSummary(summary.String())
}
