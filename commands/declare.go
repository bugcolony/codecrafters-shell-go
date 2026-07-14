package commands

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

const KeyValuePairRegex = "^_?([a-zA-Z_]*)=([^=\\s]*)$"

type VariableRegistry interface {
	Get(name string) (string, bool)
	Set(name string, value string)
}
type Declare struct {
	variables VariableRegistry
}

func (d *Declare) Name() string {
	return "declare"
}

func (d *Declare) Execute(args []string, out io.Writer, errOut io.Writer) bool {
	pFlag := ParseFlag(args, "-p", 1)

	if pFlag != nil {
		v, ok := d.variables.Get(pFlag[0])

		if !ok {
			fmt.Fprintf(out, "declare: %s: not found\n", pFlag[0])

			return true
		}

		fmt.Fprintf(out, "declare -- %s=%s\n", pFlag[0], v)

		return true
	}

	reg, err := regexp.Compile(KeyValuePairRegex)

	if err != nil {
		fmt.Println(err)
		return true
	}

	for _, pair := range args {
		if reg.MatchString(pair) {
			key, value, ok := strings.Cut(pair, "=")

			if ok {
				d.variables.Set(key, value)

				return true
			}
		} else {
			fmt.Fprintf(out, "%s: `%s': not a valid indentifier\n", d.Name(), pair)
		}
	}

	return true
}
