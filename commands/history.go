package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

type HistorySource interface {
	List() []string
	Push(lines ...string)
}

type History struct {
	source HistorySource
}

func (h *History) Name() string {
	return "history"
}

func (h *History) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	var offset int
	var fromFileList []string
	list := h.source.List()

	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])

		if err == nil && n > 0 {
			offset = len(list) - n
		}
	}

	if len(args) == 2 {
		rFlag := ParseFlag(args, "-r", 1)

		if rFlag != nil {
			f, err := os.Open(rFlag[0])

			if err != nil {
				return true
			}

			scanner := bufio.NewScanner(f)

			for scanner.Scan() {
				fromFileList = append(fromFileList, scanner.Text())
			}

			h.source.Push(fromFileList...)

			return true
		}
	}

	for i := offset; i < len(list); i++ {
		fmt.Fprintf(out, "%5d  %s\n", i+1, list[i])
	}

	return true
}
