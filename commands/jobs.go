package commands

import "io"

type Jobs struct{}

func (j Jobs) Name() string {
	return "jobs"
}

func (j Jobs) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	return true
}
