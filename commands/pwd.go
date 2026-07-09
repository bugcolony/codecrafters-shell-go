package commands

import (
	"fmt"
	"io"
	"os"
)

type Pwd struct{}

func (p Pwd) Name() string {
	return "pwd"
}

func (p Pwd) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	dir, err := os.Getwd()

	if err != nil {
		fmt.Fprintln(errOut, err)
	}

	fmt.Fprintln(out, dir)

	return true
}
