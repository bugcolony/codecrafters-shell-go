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
	
	for _, proc := range j.Process.List() {
		indicator = " "

		switch proc.Id {
		case j.Process.CurrentProcess:
			indicator = "+"
		case j.Process.PreviousProcess:
			indicator = "-"
		default:
			indicator = " "
		}

		fmt.Printf("[%d]%s %-24s %s\n", proc.Id, indicator, proc.State, proc.Command)
	}

	return true
}
