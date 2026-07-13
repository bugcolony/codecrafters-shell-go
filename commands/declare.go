package commands

import "io"

type Declare struct{}

func (d Declare) Name() string {
	return "declare"
}

func (d Declare) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	return true
}
