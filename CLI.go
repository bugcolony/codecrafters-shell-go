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
	for {
		fmt.Fprint(cli.out, "$ ")

		inputCommand := cli.ReadLine()

		//if inputCommand == "exit" {
		//	return
		//}

		fmt.Fprintf(cli.out, "%s: command not found\n", inputCommand)
	}
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return cli.in.Text()
}
