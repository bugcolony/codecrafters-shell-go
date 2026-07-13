package commands

import "io"

type History struct{}

func (h History) Name() string {
	return "history"
}

func (h History) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	return true
}
