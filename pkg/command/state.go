package command

import (
	"errors"
	"flag"
	"io"
	"slices"
	"strings"
)

type StateRm struct {
	Project   string
	Workspace string
	Addresses []string
	DryRun    bool
}

var _ Command = (*StateRm)(nil)

func (s *StateRm) Type() Type {
	return StateType
}

func parseStateCommand(args []string) (*StateRm, error) {
	var (
		cmds []string
		opts []string
	)
	n := slices.Index(args, dash)
	if n == -1 {
		cmds = args[2:]
	} else {
		cmds = args[2:n]
		opts = args[n+1:]
	}
	state := &StateRm{}
	flagSet := flag.NewFlagSet("state", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&state.Project, "p", "", "")
	flagSet.StringVar(&state.Project, "project", "", "")
	flagSet.StringVar(&state.Workspace, "w", "", "")
	flagSet.StringVar(&state.Workspace, "workspace", "", "")
	if err := flagSet.Parse(cmds); err != nil {
		return nil, err
	}
	arr := flagSet.Args()
	if len(arr) < 2 {
		return nil, errors.New("invalid state command")
	}
	switch strings.ToLower(arr[0]) {
	case "rm":
	default:
		return nil, errors.New("invalid state sub")
	}
	state.Addresses = arr[1:]
	flagSet = flag.NewFlagSet("opts", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.BoolVar(&state.DryRun, "dry-run", false, "")
	if err := flagSet.Parse(opts); err != nil {
		return nil, err
	}
	return state, nil
}
