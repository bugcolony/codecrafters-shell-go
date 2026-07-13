package commands

import (
	"io"
	"slices"
)

type Command interface {
	Name() string
	Execute(args []string, out io.Writer, errOut io.Writer) bool
}

func ParseFlag(args []string, flag string, count int) []string {
	idx := slices.Index(args, flag)
	start := idx + 1
	end := start + count

	if idx == -1 || end > len(args) {
		return nil
	}

	return args[start:end]
}
