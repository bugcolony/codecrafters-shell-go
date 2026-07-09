package commands

import "io"

type Command interface {
	Name() string
	Execute(args []string, out io.Writer, errOut io.Writer) bool
}
