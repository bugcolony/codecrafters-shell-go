package commands

import (
	"fmt"
	"io"
)

type Jobs struct {
	Process *ProcessTable
}

func (j *Jobs) Name() string {
	return "jobs"
}

func (j *Jobs) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	indicator := " "

	list := j.Process.List()

	for idx, proc := range list {
		indicator = " "

		switch idx {
		case len(list) - 1:
			indicator = "+"
		case len(list) - 2:
			indicator = "-"
		default:
			indicator = " "
		}

		fmt.Printf("[%d]%s %-24s %s\n", proc.Id, indicator, proc.State, proc.Command)
	}

	return true
}
