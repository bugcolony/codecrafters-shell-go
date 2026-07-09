package commands

import "io"

type Exit struct{}

func (e Exit) Name() string {
	return "exit"
}

func (e Exit) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	return false
}
