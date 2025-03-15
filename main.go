package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/yu-icchi/mu/pkg/action"
	"github.com/yu-icchi/mu/pkg/app"
	"github.com/yu-icchi/mu/pkg/github"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("mu (version=%s, commit=%s, date=%s)", version, commit, date)) // nolint: gosimple

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	owner := action.Owner()
	repo := action.Repo()
	token := action.Input("github_token")
	if token == "" {
		action.Failed("invalid github_token")
	}
	configPath := action.Input("config_path")
	if configPath == "" {
		action.Failed("invalid config_path")
	}
	uploadArtifactVersion := action.Input("upload_artifact_version")
	uploadArtifactDir := action.Input("upload_artifact_dir")
	defaultTerraformVersion := action.Input("default_terraform_version")
	disableSummaryLog, err := strconv.ParseBool(action.Input("disable_summary_log"))
	if err != nil {
		action.Failed("invalid disable_summary_log")
	}
	allowCommands := strings.Split(strings.ToLower(action.Input("allow_commands")), ",")
	emojiReaction := action.Input("emoji_reaction")
	gh, err := github.New(ctx, token, owner, repo)
	if err != nil {
		action.Failed("failed to setup github client")
	}

	params := &app.Params{
		Github:                  gh,
		ConfigPath:              configPath,
		DefaultTerraformVersion: defaultTerraformVersion,
		UploadArtifactVersion:   uploadArtifactVersion,
		UploadArtifactDir:       uploadArtifactDir,
		AllowCommands:           allowCommands,
		DisableSummaryLog:       disableSummaryLog,
		EmojiReaction:           emojiReaction,
		Release: &app.Release{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	mu := app.New(params)
	if err := mu.Execute(ctx); err != nil {
		action.Failed(err.Error())
	}
}
