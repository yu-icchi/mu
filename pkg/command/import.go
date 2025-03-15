package command

import (
	"flag"
	"io"
	"slices"
)

type Import struct {
	Project   string
	Workspace string
	Address   string
	ID        string
	Vars      TerraformVars
	VarFiles  TerraformVarFiles
}

var _ Command = (*Import)(nil)

func (i *Import) Type() Type {
	return ImportType
}

func parseImportCommand(args []string) (*Import, error) {
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
	importCmd := &Import{}
	flagSet := flag.NewFlagSet("import", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&importCmd.Project, "p", "", "")
	flagSet.StringVar(&importCmd.Project, "project", "", "")
	flagSet.StringVar(&importCmd.Workspace, "w", "", "")
	flagSet.StringVar(&importCmd.Workspace, "workspace", "", "")
	if err := flagSet.Parse(cmds); err != nil {
		return nil, err
	}
	importCmd.Address = flagSet.Arg(0)
	importCmd.ID = flagSet.Arg(1)
	flagSet = flag.NewFlagSet("opts", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Var(&importCmd.Vars, "var", "")
	flagSet.Var(&importCmd.VarFiles, "var-file", "")
	if err := flagSet.Parse(opts); err != nil {
		return nil, err
	}
	return importCmd, nil
}
