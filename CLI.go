package codecrafters_shell_go

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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

		inputLine := strings.Split(cli.ReadLine(), " ")

		if len(inputLine) == 0 {
			continue
		}

		cmd := inputLine[0]

		switch cmd {
		case "exit":
			return
		case "echo":
			fmt.Fprintln(cli.out, strings.Join(inputLine[1:], " "))
		default:
			fmt.Fprintf(cli.out, "%s: command not found\n", cmd)
		}

	}
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return cli.in.Text()
}
