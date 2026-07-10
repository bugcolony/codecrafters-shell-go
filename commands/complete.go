package commands

import (
	"fmt"
	"io"
	"slices"
)

type CompletionReg interface {
	Set(command string, script string)
	Get(command string) (string, bool)
	Remove(command string)
}

type Complete struct {
	CompleteRegistry CompletionReg
}

func (c *Complete) Name() string {
	return "complete"
}

func (c *Complete) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	pFlag := parseFlag(args, "-p", 1)
	cFlag := parseFlag(args, "-C", 2)
	rFlag := parseFlag(args, "-r", 1)

	if pFlag != nil {
		if reg, ok := c.CompleteRegistry.Get(pFlag[0]); ok {
			fmt.Fprintf(out, "complete -C '%s' %s\n", reg, pFlag[0])
		} else {
			fmt.Fprintf(out, "complete: %s: no completion specification\n", pFlag[0])
		}
	}

	if cFlag != nil {
		c.CompleteRegistry.Set(cFlag[1], cFlag[0])
	}

	if rFlag != nil {
		c.CompleteRegistry.Remove(rFlag[0])
	}

	return true
}

func parseFlag(args []string, flag string, count int) []string {
	idx := slices.Index(args, flag)
	start := idx + 1
	end := start + count

	if idx == -1 || end > len(args) {
		return nil
	}

	return args[start:end]
}
