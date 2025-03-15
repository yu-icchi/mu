package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"
)

type Type string

const (
	PlanType   Type = "plan"
	ApplyType  Type = "apply"
	UnlockType Type = "unlock"
	HelpType   Type = "help"
	ImportType Type = "import"
	StateType  Type = "state"
)

const (
	muCmd = "mu"
	dash  = "--"
)

var ErrInvalidCommand = errors.New("mu: invalid mu command")

type Command interface {
	Type() Type
}

type TerraformVars []string

func (t *TerraformVars) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *TerraformVars) Set(str string) error {
	*t = append(*t, str)
	return nil
}

type TerraformVarFiles []string

func (t *TerraformVarFiles) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *TerraformVarFiles) Set(str string) error {
	*t = append(*t, str)
	return nil
}

func Parse(msg string) (Command, error) {
	msg = strings.TrimSpace(msg)
	if strings.ContainsFunc(msg, func(r rune) bool {
		return r == '\r' || r == '\n'
	}) {
		return nil, ErrInvalidCommand
	}

	args, err := shlex.Split(msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
	}
	if len(args) < 2 {
		return nil, ErrInvalidCommand
	}
	if strings.ToLower(args[0]) != muCmd {
		return nil, ErrInvalidCommand
	}

	switch Type(strings.ToLower(args[1])) {
	case PlanType:
		cmd, err := parsePlanCommand(args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
		}
		return cmd, nil
	case ApplyType:
		cmd, err := parseApplyCommand(args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
		}
		return cmd, nil
	case UnlockType:
		cmd, err := parseUnlockCommand(args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
		}
		return cmd, nil
	case HelpType:
		cmd := parseHelpCommand()
		return cmd, nil
	case ImportType:
		cmd, err := parseImportCommand(args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
		}
		return cmd, nil
	case StateType:
		cmd, err := parseStateCommand(args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidCommand, err)
		}
		return cmd, nil
	default:
		return nil, ErrInvalidCommand
	}
}
