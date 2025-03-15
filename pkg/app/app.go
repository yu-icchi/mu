package app

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/yu-icchi/mu/pkg/action"
	"github.com/yu-icchi/mu/pkg/archive"
	"github.com/yu-icchi/mu/pkg/command"
	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/log"
	"github.com/yu-icchi/mu/pkg/terraform"
)

type Release struct {
	Version string
	Commit  string
	Date    string
}

type App struct {
	terraform               terraform.Terraform
	github                  github.Github
	action                  *action.Action
	archiver                archive.Archive
	configPath              string
	defaultTerraformVersion string
	uploadArtifactVersion   string
	uploadArtifactDir       string
	allowCommands           []string
	logger                  log.Logger
	disableSummaryLog       bool
	emojiReaction           string
	release                 *Release
}

type Params struct {
	Github                  github.Github
	ConfigPath              string
	DefaultTerraformVersion string
	UploadArtifactVersion   string
	UploadArtifactDir       string
	AllowCommands           []string
	DisableSummaryLog       bool
	EmojiReaction           string
	Release                 *Release
}

func New(params *Params) *App {
	return &App{
		terraform:               nil,
		github:                  params.Github,
		action:                  action.New(os.Stdout),
		archiver:                archive.NewZipArchiver(),
		configPath:              params.ConfigPath,
		defaultTerraformVersion: params.DefaultTerraformVersion,
		uploadArtifactVersion:   params.UploadArtifactVersion,
		uploadArtifactDir:       params.UploadArtifactDir,
		allowCommands:           params.AllowCommands,
		logger:                  log.New(os.Stdout),
		disableSummaryLog:       params.DisableSummaryLog,
		emojiReaction:           params.EmojiReaction,
		release:                 params.Release,
	}
}

func (a *App) Execute(ctx context.Context) error {
	event, err := a.github.Event()
	if err != nil {
		return err
	}
	switch e := event.(type) {
	case *github.PullRequestEvent:
		return a.executePullRequestEvent(ctx, e)
	case *github.IssueCommentEvent:
		return a.executeIssueCommentEvent(ctx, e)
	default:
		return nil
	}
}

func (a *App) executePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) error {
	eventAction := event.GetAction()
	switch eventAction {
	case github.Opened, github.Synchronize, github.Reopened, github.Closed:
	default:
		return nil
	}

	cfg, err := config.Load(a.configPath, config.WithDefaultTerraformVersion(a.defaultTerraformVersion))
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	prNum := event.Number()
	pr, err := a.github.GetPullRequest(ctx, prNum)
	if err != nil {
		return err
	}
	if eventAction == github.Closed {
		return a.executeUnlock(ctx, prNum, cfg, &command.Unlock{})
	}
	return a.executeTerraformAutoPlan(ctx, prNum, pr.HeadSHA, cfg)
}

func (a *App) executeIssueCommentEvent(ctx context.Context, event *github.IssueCommentEvent) error {
	if event.GetAction() != github.Created {
		return nil
	}

	comment := event.GetComment()
	muCmd, err := command.Parse(comment.GetBody())
	if err != nil {
		return nil
	}

	if a.emojiReaction != "" {
		if err := a.github.CreateIssueCommentReaction(ctx, comment.GetID(), a.emojiReaction); err != nil {
			return err
		}
	}

	prNum := event.Number()

	if muCmd.Type() != command.HelpType && !slices.Contains(a.allowCommands, string(muCmd.Type())) {
		msg := a.unknownCommandMessage(string(muCmd.Type()), a.allowCommands)
		if err := a.github.CreateIssueComment(ctx, prNum, msg); err != nil {
			return err
		}
		return nil
	}

	pr, err := a.github.GetPullRequest(ctx, prNum)
	if err != nil {
		return err
	}
	if !pr.IsMergeable() {
		return fmt.Errorf("conflict: %s", pr.MergeableState)
	}
	sha := pr.HeadSHA

	cfg, err := config.Load(a.configPath, config.WithDefaultTerraformVersion(a.defaultTerraformVersion))
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	switch cmd := muCmd.(type) {
	case *command.Plan:
		return a.executeTerraformPlan(ctx, prNum, sha, cfg, cmd)
	case *command.Apply:
		return a.executeTerraformApply(ctx, prNum, sha, cfg, cmd)
	case *command.Unlock:
		return a.executeUnlock(ctx, prNum, cfg, cmd)
	case *command.Help:
		return a.executeHelp(ctx, prNum)
	case *command.Import:
		return a.executeTerraformImport(ctx, prNum, sha, cfg, cmd)
	case *command.StateRm:
		return a.executeTerraformStateRm(ctx, prNum, sha, cfg, cmd)
	default:
		return nil
	}
}

func (a *App) executeUnlock(ctx context.Context, prNum int, cfg *config.Config, cmd *command.Unlock) error {
	projects := make(config.Projects, 0, len(cfg.Projects))
	if cmd.Project == "" {
		modifiedFiles, err := a.github.ListFiles(ctx, prNum)
		if err != nil {
			return err
		}
		for _, project := range cfg.Projects {
			if project.Plan.HasMatchedPaths(project.Dir, modifiedFiles) {
				projects = append(projects, project)
			}
		}
	} else {
		if project := cfg.GetProject(cmd.Project); project != nil {
			projects = append(projects, project)
		}
	}
	if len(projects) == 0 {
		return nil
	}

	pr, err := a.github.GetPullRequest(ctx, prNum)
	if err != nil {
		return err
	}
	if len(projects) > 1 && cmd.ForceUnlockID != "" {
		return errInvalidForceUnlock
	}

	artifactNames := make([]string, 0, len(projects))
	for _, project := range projects {
		if len(projects) == 1 && cmd.ForceUnlockID != "" {
			if err := a.tfForceUnlock(ctx, prNum, project, cmd); err != nil {
				return err
			}
		}
		if err := a.unlock(ctx, project.Name, pr); err != nil {
			return err
		}
		artifactNames = append(artifactNames, a.genArtifactName(project.Name, project.Workspace, prNum))
	}

	if err := a.github.DeleteArtifactsByNames(ctx, artifactNames); err != nil {
		return err
	}

	if err := a.deleteProgressLabel(ctx, prNum); err != nil {
		if !github.IsErrNotFound(err) {
			return err
		}
	}
	return nil
}

func (a *App) executeHelp(ctx context.Context, prNum int) error {
	return a.github.CreateIssueComment(ctx, prNum, a.helpMessage())
}
