package app

import (
	"context"
	"io"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yu-icchi/mu/pkg/action"
	archiveMock "github.com/yu-icchi/mu/pkg/archive/mock"
	githubMock "github.com/yu-icchi/mu/pkg/github/mock"
	"github.com/yu-icchi/mu/pkg/log"
	tfMock "github.com/yu-icchi/mu/pkg/terraform/mock"
)

type mock struct {
	github    *githubMock.MockGithub
	archive   *archiveMock.MockArchive
	terraform *tfMock.MockTerraform
}

func newMock(ctrl *gomock.Controller) *mock {
	return &mock{
		github:    githubMock.NewMockGithub(ctrl),
		archive:   archiveMock.NewMockArchive(ctrl),
		terraform: tfMock.NewMockTerraform(ctrl),
	}
}

func newTestAppAndMock(ctrl *gomock.Controller) (*App, *mock) {
	mock := newMock(ctrl)
	app := &App{
		terraform:               mock.terraform,
		github:                  mock.github,
		action:                  action.New(io.Discard),
		archiver:                mock.archive,
		configPath:              "",
		defaultTerraformVersion: "",
		uploadArtifactVersion:   "",
		uploadArtifactDir:       "./test-upload-artifact",
		allowCommands:           nil,
		logger:                  log.New(io.Discard),
		release: &Release{
			Version: "test-version",
			Commit:  "test-commit",
			Date:    "test-date",
		},
		disableSummaryLog: false,
		emojiReaction:     "",
	}
	return app, mock
}

type prepare func(ctx context.Context, m *mock, t *testing.T)
