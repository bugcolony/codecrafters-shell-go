package commands

import (
	"fmt"
	"io"
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
	pFlag := ParseFlag(args, "-p", 1)
	cFlag := ParseFlag(args, "-C", 2)
	rFlag := ParseFlag(args, "-r", 1)

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
