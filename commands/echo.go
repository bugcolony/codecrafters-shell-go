package commands

import (
	"fmt"
	"io"
	"strings"
)

type Echo struct{}

func (e Echo) Name() string {
	return "echo"
}

func (e Echo) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	fmt.Fprintf(out, "%s\n", strings.Join(args, " "))

	return true
}
