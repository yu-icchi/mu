package command

import (
	"flag"
	"io"
	"slices"
)

type Unlock struct {
	Project       string
	Workspace     string
	ForceUnlockID string
}

var _ Command = (*Unlock)(nil)

func (u *Unlock) Type() Type {
	return UnlockType
}

func parseUnlockCommand(args []string) (*Unlock, error) {
	var cmds []string
	n := slices.Index(args, dash)
	if n == -1 {
		cmds = args[2:]
	} else {
		cmds = args[2:n]
	}
	unlock := &Unlock{}
	flagSet := flag.NewFlagSet("unlock", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&unlock.Project, "p", "", "")
	flagSet.StringVar(&unlock.Project, "project", "", "")
	flagSet.StringVar(&unlock.Workspace, "w", "", "")
	flagSet.StringVar(&unlock.Workspace, "workspace", "", "")
	flagSet.StringVar(&unlock.ForceUnlockID, "force-unlock", "", "")
	if err := flagSet.Parse(cmds); err != nil {
		return nil, err
	}
	return unlock, nil
}
