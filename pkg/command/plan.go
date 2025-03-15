package command

import (
	"flag"
	"io"
	"slices"
)

type Plan struct {
	Project   string
	Workspace string
	Vars      TerraformVars
	VarFiles  TerraformVarFiles
	Destroy   bool
}

var _ Command = (*Plan)(nil)

func (p *Plan) Type() Type {
	return PlanType
}

func parsePlanCommand(args []string) (*Plan, error) {
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
	plan := &Plan{}
	flagSet := flag.NewFlagSet("plan", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&plan.Project, "p", "", "")
	flagSet.StringVar(&plan.Project, "project", "", "")
	flagSet.StringVar(&plan.Workspace, "w", "", "")
	flagSet.StringVar(&plan.Workspace, "workspace", "", "")
	if err := flagSet.Parse(cmds); err != nil {
		return nil, err
	}
	flagSet = flag.NewFlagSet("opts", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Var(&plan.Vars, "var", "")
	flagSet.Var(&plan.VarFiles, "var-file", "")
	flagSet.BoolVar(&plan.Destroy, "destroy", false, "")
	if err := flagSet.Parse(opts); err != nil {
		return nil, err
	}
	return plan, nil
}
