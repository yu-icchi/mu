package command

import (
	"flag"
	"io"
)

type Apply struct {
	Project   string
	Workspace string
}

var _ Command = (*Apply)(nil)

func (a *Apply) Type() Type {
	return ApplyType
}

func parseApplyCommand(args []string) (*Apply, error) {
	apply := &Apply{}
	flagSet := flag.NewFlagSet("apply", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&apply.Project, "p", "", "")
	flagSet.StringVar(&apply.Project, "project", "", "")
	flagSet.StringVar(&apply.Workspace, "w", "", "")
	flagSet.StringVar(&apply.Workspace, "workspace", "", "")
	if err := flagSet.Parse(args[2:]); err != nil {
		return nil, err
	}
	return apply, nil
}
