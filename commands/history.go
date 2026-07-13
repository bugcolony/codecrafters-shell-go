package commands

import (
	"fmt"
	"io"
	"strconv"
)

type HistoryLister interface {
	List() []string
}

type History struct {
	source HistoryLister
}

func (h *History) Name() string {
	return "history"
}

func (h *History) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	var offset int
	list := h.source.List()

	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])

		if err == nil && n > 0 {
			offset = len(list) - n
		}
	}

	for i := offset; i < len(list); i++ {
		fmt.Fprintf(out, "%5d %s\n", i+1, list[i])
	}

	return true
}
