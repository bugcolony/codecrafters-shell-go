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
	WriteToFile(filename string) error
	AppendToFile(filename string) error
}

type History struct {
	source HistorySource
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

	if len(args) == 2 {
		rFlag := ParseFlag(args, "-r", 1)
		wFlag := ParseFlag(args, "-w", 1)
		aFlag := ParseFlag(args, "-a", 1)

		if rFlag != nil {
			return h.handleFileRead(rFlag[0])
		}

		if wFlag != nil {
			return h.handleFileWrite(wFlag[0])
		}

		if aFlag != nil {
			return h.handleFileAppend(aFlag[0])
		}
	}

	for i := offset; i < len(list); i++ {
		fmt.Fprintf(out, "%5d  %s\n", i+1, list[i])
	}

	return true
}

func (h *History) handleFileRead(name string) bool {
	var fromFileList []string
	f, err := os.Open(name)

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

func (h *History) handleFileWrite(name string) bool {
	_ = h.source.WriteToFile(name)

	return true
}

func (h *History) handleFileAppend(name string) bool {
	_ = h.source.AppendToFile(name)

	return true
}
