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

func (cli *CLI) BuiltinCommands() map[string]bool {
	return map[string]bool{
		"echo": true,
		"type": true,
		"exit": true,
	}
}

func (cli *CLI) CommandExists(cmd string) bool {
	_, exist := cli.BuiltinCommands()[cmd]

	return exist
}

func (cli *CLI) printNotFound(cmd string) {
	fmt.Fprintf(cli.out, "%s: command not found\n", cmd)
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
		case "type":
			if len(inputLine) < 2 {
				continue
			}

			arg := strings.TrimSpace(inputLine[1])

			if _, exists := cli.BuiltinCommands()[arg]; exists {
				fmt.Fprintf(cli.out, "%s is a shell builtin\n", arg)
			} else {
				cli.printNotFound(arg)
			}
		default:
			cli.printNotFound(cmd)
		}

	}
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return cli.in.Text()
}
