package commands

import (
	"fmt"
	"io"
	"os"
)

const (
	HomePathAlias = "~"
)

type Cd struct{}

func (c Cd) Name() string {
	return "cd"
}

func (c Cd) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	if len(args) < 1 {
		return true
	}

	pathArg := args[0]

	if pathArg == HomePathAlias {
		pathArg = os.Getenv("HOME")
	}

	if err := os.Chdir(pathArg); err != nil {
		fmt.Fprintf(out, "cd: %s: No such file or directory\n", pathArg)
	}

	return true
}
