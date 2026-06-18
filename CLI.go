package codecrafters_shell_go

import (
	"bufio"
	"fmt"
	"io"
)

type CLI struct {
	in  *bufio.Scanner
	out io.Writer
}

func NewCLI(in io.Reader, out io.Writer) *CLI {
	return &CLI{
		in:  bufio.NewScanner(in),
		out: out,
	}
}

func (cli *CLI) Run() {
	fmt.Fprint(cli.out, "$")

	cli.in.Scan()

	inputCommand := cli.in.Text()

	fmt.Fprintf(cli.out, "%s: command not found\n", inputCommand)
}
