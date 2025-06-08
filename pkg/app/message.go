package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/yu-icchi/mu/pkg/config"
	"github.com/yu-icchi/mu/pkg/github"
	"github.com/yu-icchi/mu/pkg/terraform"
)

const (
	muInitMeta  = "<!-- mu:init -->"
	muPlanMeta  = "<!-- mu:plan -->"
	muApplyMeta = "<!-- mu:apply -->"
)

func (a *App) unknownCommandMessage(commandType string, allowCommands []string) string {
	msg := new(strings.Builder)
	msg.WriteString("```\n")
	msg.WriteString(fmt.Sprintf("Error: unknown command %q.\n", commandType))
	msg.WriteString("Run 'mu help' for usage.\n")
	msg.WriteString(fmt.Sprintf("Available commands: %s\n", strings.Join(allowCommands, ", ")))
	msg.WriteString("```\n")
	return msg.String()
}

func (a *App) helpMessage() string {
	const msg = `Mu
Terraform Pull Request Automation

Usage:
  mu <command> [options] -- [terraform options]

Examples:
  # show atlantis help
  mu help

  # run plan in the project passing the -var flag to terraform
  mu plan -p <project> -- -var name=test

  # apply the plan for the project
  mu apply -p <project>

Commands:
  plan     Runs 'terraform plan' for the changes in this pull request.
           To plan a specific project, use the -p flags.

  apply    Runs 'terraform apply' on all unapplied plans from this pull request.
           To only apply a specific plan, use the -p flags.

  unlock   Removes all mu locks and discards all plans for this pull request.

  help     View help.

`
	return "```\n" + msg + "```\n"
}

func (a *App) initFailedMessage(cfg *config.Project, out *terraform.Output) string {
	msg := new(strings.Builder)
	msg.WriteString(muInitMeta)
	msg.WriteString("\n:x: **Init Failed**\n")
	msg.WriteString(fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", cfg.Name, cfg.Dir, cfg.Workspace))
	cautionResult := a.formatMarkdownAlert("CAUTION", out.Result)
	msg.WriteString(cautionResult)
	return msg.String()
}

func (a *App) formatMarkdownAlert(alert, text string) string {
	if text == "" {
		return ""
	}
	ret := new(strings.Builder)
	ret.WriteString(fmt.Sprintf("> [!%s]\n", strings.ToUpper(alert)))
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		ret.WriteString("> ")
		ret.WriteString(scanner.Text())
		ret.WriteString("\n")
	}
	return ret.String()
}

func (a *App) splitMessages(reader io.Reader) []string {
	const (
		startDetails = "<details>"
		endDetails   = "</details>"
		startSummary = "<summary>"
		endSummary   = "</summary>"
		codeBlock    = "```"
		diff         = "diff"
		warning      = "> [!WARNING]"
		size         = github.MaxCommentLen - 5536
	)

	var msgs []string
	msg := new(strings.Builder)
	var (
		isDetails, isCodeBlock, isWarning, isDiff bool
		summaryTitle, codeBlockSpace              string
		count                                     int
	)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		switch {
		case strings.HasPrefix(text, startDetails+startSummary):
			isDetails = true
			idx := strings.Index(text, endSummary)
			summaryTitle = text[len(startDetails+startSummary):idx]
		case strings.HasPrefix(text, endDetails):
			isDetails = false
		case strings.Contains(text, codeBlock) && !isCodeBlock:
			isCodeBlock = true
			idx := strings.Index(text, codeBlock)
			codeBlockSpace = text[0:idx]
			if strings.Contains(text, codeBlock+diff) {
				isDiff = true
			}
		case strings.Contains(text, codeBlock) && isCodeBlock:
			isCodeBlock = false
			if isDiff {
				isDiff = false
			}
		case !isWarning && strings.HasPrefix(text, warning):
			isWarning = true
		}
		if count+len(text)+1 > size {
			if isCodeBlock {
				// end code block
				msg.WriteString(codeBlockSpace)
				msg.WriteString("```\n\n")
			}
			if isDetails {
				msg.WriteString("</details>\n")
			}
			msg.WriteString("\n**Warning** Continued in next comment.\n")
			msgs = append(msgs, msg.String())
			msg.Reset()
			count = 0
			msg.WriteString("Continued from previous comment.\n\n")
			if isDetails {
				msg.WriteString(fmt.Sprintf("<details><summary>%s</summary>\n\n", summaryTitle))
			}
			if isCodeBlock {
				// start code block
				msg.WriteString(codeBlockSpace)
				msg.WriteString("```")
				if isDiff {
					msg.WriteString("diff")
				}
				msg.WriteString("\n")
			}

			if isWarning {
				msg.WriteString(warning)
				msg.WriteString("\n")
			}
		}
		msg.WriteString(text)
		msg.WriteString("\n")
		count += len(text) + 1
	}
	if msg.Len() > 0 {
		msgs = append(msgs, msg.String())
	}
	return msgs
}

func (a *App) formatDiffMarkdownChangeResult(result string) *bytes.Buffer {
	if result == "" {
		return nil
	}
	ret := new(bytes.Buffer)
	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		text := a.diffMarkdown(scanner.Text())
		ret.WriteString(text)
		ret.WriteString("\n")
	}
	return ret
}

var (
	diffKeywordRegex = regexp.MustCompile(`(?m)^( +)([-+~])`)
	diffTildeRegex   = regexp.MustCompile(`(?m)^~`)
)

func (a *App) diffMarkdown(text string) string {
	formattedTerraformOutput := diffKeywordRegex.ReplaceAllString(text, "$2$1")
	formattedTerraformOutput = diffTildeRegex.ReplaceAllString(formattedTerraformOutput, "!")
	return formattedTerraformOutput
}

func (a *App) planSucceededMessage(cfg *config.Project, out *terraform.Output) string {
	msg := new(strings.Builder)
	msg.WriteString(muPlanMeta)
	msg.WriteString("\n:white_check_mark: **Plan Result**\n")
	msg.WriteString(fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", cfg.Name, cfg.Dir, cfg.Workspace))
	msg.WriteString("\n```\n")
	msg.WriteString(out.Result)
	msg.WriteString("\n```\n\n\n")
	changeResult := a.formatDiffMarkdownChangeResult(out.ChangedResult)
	if changeResult != nil {
		msg.WriteString("<details><summary>Show Output</summary>\n\n")
		msg.WriteString("```diff\n")
		msg.WriteString(changeResult.String())
		msg.WriteString("\n```\n</details>\n\n")
	}
	msg.WriteString("**next step**\n")
	msg.WriteString("- To apply this plan, comment:\n")
	msg.WriteString("  ```\n")
	msg.WriteString(fmt.Sprintf("  mu apply -p %s\n", cfg.Name))
	msg.WriteString("  ```\n")
	msg.WriteString("- To delete this plan and lock, comment:\n")
	msg.WriteString("  ```\n")
	msg.WriteString(fmt.Sprintf("  mu unlock -p %s\n", cfg.Name))
	msg.WriteString("  ```\n")
	msg.WriteString("- To plan this project again, comment:\n")
	msg.WriteString("  ```\n")
	msg.WriteString(fmt.Sprintf("  mu plan -p %s\n", cfg.Name))
	msg.WriteString("  ```\n")
	warnResult := a.formatMarkdownAlert("WARNING", out.Warning)
	if warnResult != "" {
		msg.WriteString(warnResult)
		msg.WriteString("\n\n")
	}
	return msg.String()
}

func (a *App) planFailedMessage(cfg *config.Project, out *terraform.Output) string {
	msg := new(strings.Builder)
	msg.WriteString(muPlanMeta)
	msg.WriteString("\n:x: **Plan Failed**\n")
	msg.WriteString(fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", cfg.Name, cfg.Dir, cfg.Workspace))
	cautionResult := a.formatMarkdownAlert("CAUTION", out.Result)
	msg.WriteString(cautionResult)
	return msg.String()
}

func (a *App) applySucceededMessage(cfg *config.Project, out *terraform.Output) string {
	msg := new(strings.Builder)
	msg.WriteString(muApplyMeta)
	msg.WriteString("\n:white_check_mark: **Apply Result**\n")
	msg.WriteString(fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", cfg.Name, cfg.Dir, cfg.Workspace))
	msg.WriteString("\n```\n")
	msg.WriteString(out.Result)
	msg.WriteString("\n```\n")
	warnResult := a.formatMarkdownAlert("WARNING", out.Warning)
	if warnResult != "" {
		msg.WriteString(warnResult)
		msg.WriteString("\n\n")
	}
	return msg.String()
}

func (a *App) applyFailedMessage(cfg *config.Project, out *terraform.Output) string {
	msg := new(strings.Builder)
	msg.WriteString(muApplyMeta)
	msg.WriteString("\n:x: **Apply Failed**\n")
	msg.WriteString(fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", cfg.Name, cfg.Dir, cfg.Workspace))
	cautionResult := a.formatMarkdownAlert("CAUTION", out.Result)
	msg.WriteString(cautionResult)
	return msg.String()
}

func (a *App) forceUnlockMessage(ret *terraform.ForceUnlockOutput) string {
	msg := new(strings.Builder)
	if ret.HasError {
		msg.WriteString(":x: **Force Unlock Failed**\n")
	} else {
		msg.WriteString(":white_check_mark: **Force Unlock**\n")
	}
	msg.WriteString("\n```\n")
	msg.WriteString(ret.Result)
	msg.WriteString("\n```\n")
	return msg.String()
}

func (a *App) importMessage(
	project, address, id, log string,
) string {
	msg := new(strings.Builder)
	msg.WriteString("## mu import -p " + project)
	msg.WriteString("\n")
	msg.WriteString("**Address**:" + address)
	msg.WriteString("**Id**:" + id)
	msg.WriteString("\n```\n")
	msg.WriteString(log)
	msg.WriteString("\n```\n")
	return msg.String()
}

func (a *App) stateRmMessage(address, log string) string {
	msg := new(strings.Builder)
	msg.WriteString("### " + address)
	msg.WriteString("\n```\n")
	msg.WriteString(log)
	msg.WriteString("\n```\n")
	return msg.String()
}
