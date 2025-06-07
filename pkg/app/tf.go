package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yu-icchi/mu/pkg/action"
	"github.com/yu-icchi/mu/pkg/artifact"
	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/log"
	"github.com/yu-icchi/mu/pkg/terraform"
)

type (
	OutputProject struct {
		Name      string `json:"name"`
		Dir       string `json:"dir"`
		Workspace string `json:"workspace"`
		Mode      string `json:"mode"`
		Result    string `json:"result"`
		ActionURL string `json:"action_url"`
	}
	OutputProjects []*OutputProject
)

func (a *App) executeTerraformAutoPlan(
	ctx context.Context, prNum int, sha string, cfg *config.Config,
) error {
	projects, err := a.findAutoPlanProjects(ctx, prNum, cfg)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return nil
	}

	// Duplicate execution prevention
	if err := a.createProgressLabel(ctx, prNum, sha); err != nil {
		if github.IsErrAlreadyExists(err) {
			if err := a.outputFailedProgress(ctx, prNum); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer func() {
		if err := a.deleteProgressLabel(ctx, prNum); err != nil {
			a.logger.Error(err.Error())
		}
	}()

	outputProjects := make(OutputProjects, 0, len(projects))
	artifacts := make([]*artifact.Artifact, 0, len(projects))
	for _, project := range projects {
		out, err := a.tfPlan(ctx, prNum, sha, project, &command.Plan{})
		if err != nil {
			return err
		}
		outputProjects = append(outputProjects, &OutputProject{
			Name:      project.Name,
			Dir:       project.Dir,
			Workspace: project.Workspace,
			Mode:      "plan",
			Result:    out.result,
			ActionURL: action.RunURL(),
		})
		artifacts = append(artifacts, &artifact.Artifact{
			Name:      a.genArtifactName(project.Name, project.Workspace, prNum),
			Path:      out.path,
			Overwrite: true,
		})
	}

	outputProjectsStr, err := json.Marshal(outputProjects)
	if err != nil {
		return err
	}
	_ = a.action.Output("projects", string(outputProjectsStr))

	err = artifact.UploadArtifacts(&artifact.UploadArtifactParams{
		Version:   a.uploadArtifactVersion,
		Dir:       a.uploadArtifactDir,
		Artifacts: artifacts,
	})
	if err != nil {
		return err
	}
	_ = a.action.Output("upload_artifact", "true")
	return nil
}

func (a *App) findAutoPlanProjects(ctx context.Context, prNum int, cfg *config.Config) ([]*config.Project, error) {
	modifiedFiles, err := a.github.ListFiles(ctx, prNum)
	if err != nil {
		return nil, err
	}
	projects := make([]*config.Project, 0, len(cfg.Projects))
	for _, project := range cfg.Projects {
		if !project.Plan.Auto {
			continue
		}
		if project.Plan.HasMatchedPaths(project.Dir, modifiedFiles) {
			projects = append(projects, project)
		}
	}
	return projects, nil
}

func (a *App) executeTerraformPlan(
	ctx context.Context, prNum int, sha string, cfg *config.Config, cmd *command.Plan,
) error {
	// Duplicate execution prevention
	if err := a.createProgressLabel(ctx, prNum, sha); err != nil {
		if github.IsErrAlreadyExists(err) {
			if err := a.outputFailedProgress(ctx, prNum); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer func() {
		if err := a.deleteProgressLabel(ctx, prNum); err != nil {
			a.logger.Error(err.Error())
		}
	}()

	modifiedFiles, err := a.github.ListFiles(ctx, prNum)
	if err != nil {
		return err
	}

	projects := a.findProjectConfigs(cfg, cmd.Project, modifiedFiles)
	if len(projects) == 0 {
		const msg = "There is no project to run `mu plan` on."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	outputProjects := make(OutputProjects, 0, len(projects))
	artifacts := make([]*artifact.Artifact, 0, len(projects))
	for _, project := range projects {
		// If a specific project is specified, the terraform plan may proceed regardless of the actual changes.
		if !project.HasModifiedFiles(modifiedFiles) {
			a.logger.Info(fmt.Sprintf("not found: project=%s", cmd.Project))
			continue
		}

		out, err := a.tfPlan(ctx, prNum, sha, project, cmd)
		if err != nil {
			return err
		}
		outputProjects = append(outputProjects, &OutputProject{
			Name:      project.Name,
			Dir:       project.Dir,
			Workspace: project.Workspace,
			Mode:      "plan",
			Result:    out.result,
			ActionURL: action.RunURL(),
		})
		artifacts = append(artifacts, &artifact.Artifact{
			Name:      a.genArtifactName(project.Name, project.Workspace, prNum),
			Path:      out.path,
			Overwrite: true,
		})
	}
	if len(outputProjects) == 0 {
		const msg = "The specified project could not be found."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	outputProjectsStr, err := json.Marshal(outputProjects)
	if err != nil {
		return err
	}
	_ = a.action.Output("projects", string(outputProjectsStr))

	err = artifact.UploadArtifacts(&artifact.UploadArtifactParams{
		Version:   a.uploadArtifactVersion,
		Dir:       a.uploadArtifactDir,
		Artifacts: artifacts,
	})
	if err != nil {
		return err
	}
	_ = a.action.Output("upload_artifact", "true")
	return nil
}

func (a *App) findProjectConfigs(
	cfg *config.Config, project string, modifiedFiles []string,
) config.Projects {
	projects := make([]*config.Project, 0, len(cfg.Projects))
	if project == "" {
		for _, prj := range cfg.Projects {
			if prj.Plan.HasMatchedPaths(prj.Dir, modifiedFiles) {
				projects = append(projects, prj)
			}
		}
	} else {
		if prj := cfg.GetProject(project); prj != nil {
			projects = append(projects, prj)
		}
	}
	return projects
}

func (a *App) executeTerraformApply(
	ctx context.Context, prNum int, sha string, cfg *config.Config, cmd *command.Apply,
) error {
	// Duplicate execution prevention
	if err := a.createProgressLabel(ctx, prNum, sha); err != nil {
		if github.IsErrAlreadyExists(err) {
			if err := a.outputFailedProgress(ctx, prNum); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer func() {
		if err := a.deleteProgressLabel(ctx, prNum); err != nil {
			a.logger.Error(err.Error())
		}
	}()

	modifiedFiles, err := a.github.ListFiles(ctx, prNum)
	if err != nil {
		return err
	}
	reviews, err := a.github.ListReviews(ctx, prNum)
	if err != nil {
		return err
	}

	projects := a.findProjectConfigs(cfg, cmd.Project, modifiedFiles)
	if len(projects) == 0 {
		const msg = "There is no project to plan."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	artifactNames := make([]string, 0, len(projects))
	for _, project := range projects {
		artifactName := a.genArtifactName(project.Name, project.Workspace, prNum)
		artifactNames = append(artifactNames, artifactName)
	}
	artifacts, err := a.github.MultiGetArtifactsByNames(ctx, artifactNames)
	if err != nil {
		return err
	}

	outputProjects := make(OutputProjects, 0, len(projects))
	deleteArtifactNames := make([]string, 0, len(projects))
	for _, project := range projects {
		// If a specific project is specified, the terraform apply may proceed regardless of the actual changes.
		if !project.HasModifiedFiles(modifiedFiles) {
			a.logger.Info("Not found", log.String("project", cmd.Project))
			continue
		}

		artifactName := a.genArtifactName(project.Name, project.Workspace, prNum)
		artifact := artifacts.Get(artifactName)
		out, err := a.tfApply(ctx, prNum, sha, project, artifact, reviews)
		if err != nil {
			return err
		}
		outputProjects = append(outputProjects, &OutputProject{
			Name:      project.Name,
			Dir:       project.Dir,
			Workspace: project.Workspace,
			Mode:      "apply",
			Result:    out.result,
			ActionURL: action.RunURL(),
		})
		deleteArtifactNames = append(deleteArtifactNames, artifactName)
	}
	if len(outputProjects) == 0 {
		const msg = "The specified project could not be found."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	outputProjectsStr, err := json.Marshal(outputProjects)
	if err != nil {
		return err
	}
	_ = a.action.Output("projects", string(outputProjectsStr))

	if err := a.github.DeleteArtifactsByNames(ctx, deleteArtifactNames); err != nil {
		return err
	}
	return nil
}

func (a *App) executeTerraformImport(
	ctx context.Context, prNum int, sha string, cfg *config.Config, cmd *command.Import,
) error {
	// Duplicate execution prevention
	if err := a.createProgressLabel(ctx, prNum, sha); err != nil {
		if github.IsErrAlreadyExists(err) {
			if err := a.outputFailedProgress(ctx, prNum); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer func() {
		if err := a.deleteProgressLabel(ctx, prNum); err != nil {
			a.logger.Error(err.Error())
		}
	}()

	modifiedFiles, err := a.github.ListFiles(ctx, prNum)
	if err != nil {
		return err
	}

	projects := a.findProjectConfigs(cfg, cmd.Project, modifiedFiles)
	if len(projects) != 1 {
		const msg = "Please limit to one target project."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	if err := a.tfImport(ctx, prNum, projects[0], cmd); err != nil {
		return err
	}
	return nil
}

func (a *App) executeTerraformStateRm(
	ctx context.Context, prNum int, sha string, cfg *config.Config, cmd *command.StateRm,
) error {
	// Duplicate execution prevention
	if err := a.createProgressLabel(ctx, prNum, sha); err != nil {
		if github.IsErrAlreadyExists(err) {
			if err := a.outputFailedProgress(ctx, prNum); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer func() {
		if err := a.deleteProgressLabel(ctx, prNum); err != nil {
			a.logger.Error(err.Error())
		}
	}()

	modifiedFiles, err := a.github.ListFiles(ctx, prNum)
	if err != nil {
		return err
	}

	projects := a.findProjectConfigs(cfg, cmd.Project, modifiedFiles)
	if len(projects) != 1 {
		const msg = "Please limit to one target project."
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	if err := a.tfStateRm(ctx, prNum, projects[0], cmd); err != nil {
		return err
	}
	return nil
}

func (a *App) genTerraform(cfg *config.Project) terraform.Terraform {
	if a.terraform != nil {
		return a.terraform
	}
	version := cfg.Terraform.GetVersion()
	if version == "" {
		version = terraform.LatestVersion
	}
	return terraform.New(&terraform.Params{
		Version:  version,
		WorkDir:  cfg.Dir,
		ExecPath: cfg.Terraform.GetExecPath(),
	})
}

func (a *App) genPlanFilename(name, workspace string, prNum int) string {
	name = strings.ReplaceAll(name, "/", "::")
	return fmt.Sprintf("%s_%s_%d.tfplan", name, workspace, prNum)
}

func (a *App) genArtifactName(name, workspace string, prNum int) string {
	name = strings.ReplaceAll(name, "/", "::")
	return fmt.Sprintf("mu_%s_%s_%d", name, workspace, prNum)
}

func (a *App) genProgressLabel(prNum int) string {
	return fmt.Sprintf("mu_in_progress_%d", prNum)
}

func (a *App) createProgressLabel(ctx context.Context, prNum int, sha string) error {
	label := a.genProgressLabel(prNum)
	var desc string
	if sha != "" {
		desc = fmt.Sprintf("commit: %s", sha)
	}
	if err := a.github.CreateLabel(ctx, label, desc, ""); err != nil {
		return err
	}
	return a.github.AddPullRequestLabels(ctx, prNum, []string{label})
}

func (a *App) deleteProgressLabel(ctx context.Context, prNum int) error {
	label := a.genProgressLabel(prNum)
	return a.github.DeleteLabel(ctx, label)
}

func (a *App) outputFailedProgress(ctx context.Context, prNum int) error {
	label := a.genProgressLabel(prNum)
	msg := fmt.Sprintf("Error: The operation was canceled because #%d is currently in progress. Please remove the %q label to retry.", prNum, label)
	return a.github.CreateIssueComment(ctx, prNum, msg)
}

func (a *App) outputInitFailedSummary(cfg *config.Project, log string) {
	summary := new(strings.Builder)
	summary.WriteString(fmt.Sprintf("## %s\n\n", cfg.Name))
	summary.WriteString(":x: **Init Failed**\n")
	summary.WriteString(fmt.Sprintf("project=%s workspace=%s\n", cfg.Name, cfg.Workspace))
	summary.WriteString("<details><summary>Show Output</summary>\n")
	summary.WriteString("\n```\n")
	summary.WriteString(log)
	summary.WriteString("\n```\n")
	summary.WriteString("</details>\n")
	_ = a.action.AddStepSummary(summary.String())
}

func (a *App) outputInitFailedResult(
	ctx context.Context, prNum int, cfg *config.Project, out *terraform.Output,
) error {
	comment := a.initFailedMessage(cfg, out)
	messages := a.splitMessages(strings.NewReader(comment))
	for _, msg := range messages {
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
	}
	return nil
}
