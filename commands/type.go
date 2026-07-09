package commands

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Type struct {
	Commands *Registry
}

func (t *Type) Name() string {
	return "type"
}

func (t *Type) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	if len(args) < 1 {
		return true
	}

	typeCmd := strings.TrimSpace(args[0])

	if _, exists := t.Commands.Get(typeCmd); exists {
		fmt.Fprintf(out, "%s is a shell builtin\n", typeCmd)
	} else {
		if lookup, err := exec.LookPath(typeCmd); err == nil {
			fmt.Fprintf(out, "%s is %s\n", typeCmd, lookup)
		} else {
			fmt.Fprintf(out, "%s: not found\n", typeCmd)
		}
	}

	return true
}
