package action

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

func Input(name string) string {
	const prefix = "INPUT_"
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, " ", "_")
	return os.Getenv(prefix + name)
}

func Failed(msg string) {
	const failure = 1
	_, _ = fmt.Fprintf(os.Stderr, "::error::%s\n", msg)
	os.Exit(failure)
}

func Repo() string {
	repo := os.Getenv("GITHUB_REPOSITORY")
	repo, _ = strings.CutPrefix(repo, Owner()+"/")
	return repo
}

func Owner() string {
	return os.Getenv("GITHUB_REPOSITORY_OWNER")
}

func RunURL() string {
	repo := os.Getenv("GITHUB_REPOSITORY")
	id := os.Getenv("GITHUB_RUN_ID")
	if repo == "" || id == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/actions/runs/%s", repo, id)
}

func LabelURL(label string) string {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" || label == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/labels/%s", repo, label)
}

type Action struct {
	stdout            io.Writer
	outputLocker      sync.Mutex
	stepSummaryLocker sync.Mutex
}

func New(stdout io.Writer) *Action {
	return &Action{
		stdout:            stdout,
		outputLocker:      sync.Mutex{},
		stepSummaryLocker: sync.Mutex{},
	}
}

func (a *Action) Output(key, value string) error {
	const envOutput = "GITHUB_OUTPUT"
	path := os.Getenv(envOutput)
	if path == "" {
		return nil
	}
	a.outputLocker.Lock()
	defer a.outputLocker.Unlock()
	return a.writeFile(path, fmt.Sprintf("%s=%s\n", key, value))
}

func (a *Action) AddStepSummary(msg string) error {
	const envStepSummary = "GITHUB_STEP_SUMMARY"
	path := os.Getenv(envStepSummary)
	if path == "" {
		return nil
	}
	a.stepSummaryLocker.Lock()
	defer a.stepSummaryLocker.Unlock()
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		text := scanner.Text()
		err := a.writeFile(path, text+"\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Action) writeFile(path, msg string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.WriteString(msg)
	return err
}

func (a *Action) Group(title, body string) {
	a.StartGroup(title)
	_, _ = fmt.Fprintln(a.stdout, body)
	a.EndGroup()
}

func (a *Action) StartGroup(title string) {
	_, _ = fmt.Fprintf(a.stdout, "::group::%s\n", title)
}

func (a *Action) EndGroup() {
	_, _ = fmt.Fprintln(a.stdout, "::endgroup::")
}
