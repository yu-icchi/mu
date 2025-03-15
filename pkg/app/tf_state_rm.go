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

type state struct {
	address string
	log     string
}

func (a *App) tfStateRm(ctx context.Context, prNum int, cfg *config.Project, cmd *command.StateRm) error {
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

	a.action.StartGroup(fmt.Sprintf("mu init --project %s --workspace %s", cfg.Name, cfg.Workspace))
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

	msg := new(strings.Builder)
	for _, address := range cmd.Addresses {
		a.action.StartGroup(fmt.Sprintf("mu state --project %s --workspace %s rm %s", cfg.Name, cfg.Workspace, address))
		stateRmRet, err := tf.StateRm(ctx, &terraform.StateRmParams{
			Address: address,
			DryRun:  cmd.DryRun,
		}, terraform.WithStream(os.Stdout))
		_, _ = fmt.Fprintln(os.Stdout)
		a.action.EndGroup()
		if err != nil {
			return err
		}
		if !a.disableSummaryLog {
			a.outputStateRmRawLog(cfg, stateRmRet.Result, address)
		}
		msg.WriteString(a.stateRmMessage(address, stateRmRet.Result))
		msg.WriteString("\n")
	}

	if err := a.github.CreateIssueComment(ctx, prNum, msg.String()); err != nil {
		return err
	}
	return nil
}

func (a *App) outputStateRmRawLog(cfg *config.Project, log, address string) {
	summary := new(strings.Builder)
	summary.WriteString("## mu state rm\n\n")
	summary.WriteString(fmt.Sprintf("**address: %s**\n", address))
	summary.WriteString(fmt.Sprintf("workspace: `%s` workspace: `%s`", cfg.Name, cfg.Workspace))
	summary.WriteString("<details><summary>Show Output</summary>\n")
	summary.WriteString("\n```\n")
	summary.WriteString(log)
	summary.WriteString("\n```\n")
	summary.WriteString("</details>\n")
	_ = a.action.AddStepSummary(summary.String())
}
