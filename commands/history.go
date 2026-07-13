package commands

import (
	"fmt"
	"io"
)

type HistoryLister interface {
	List() [][]byte
}

type History struct {
	source HistoryLister
}

func (h *History) Name() string {
	return "history"
}

func (h *History) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	for _, line := range h.source.List() {
		fmt.Fprintln(out, string(line))
	}

	return true
}
