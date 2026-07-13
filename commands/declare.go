package commands

import (
	"fmt"
	"io"
)

type Declare struct{}

func (d Declare) Name() string {
	return "declare"
}

func (d Declare) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	pFlag := ParseFlag(args, "-p", 1)

	if pFlag != nil {
		fmt.Fprintf(out, "declare: %s: not found\n", pFlag[0])
	}

	return true
}
