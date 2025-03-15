package command

type Help struct{}

var _ Command = (*Help)(nil)

func (h *Help) Type() Type {
	return HelpType
}

func parseHelpCommand() *Help {
	return &Help{}
}
