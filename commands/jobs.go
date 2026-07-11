package commands

import (
	"fmt"
	"io"
)

type Jobs struct {
	Process *ProcessList
}

func (j *Jobs) Name() string {
	return "jobs"
}

func (j *Jobs) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	idx := 0
	last := ""

	for _, proc := range j.Process.List() {
		idx++

		if idx == len(j.Process.List()) {
			last = "+"
		}

		fmt.Printf("[%d]%s %-24s %s\n", idx, last, proc.State, proc.Command)
	}

	return true
}
