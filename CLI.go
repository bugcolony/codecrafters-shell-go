package codecrafters_shell_go

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const (
	homePathAlias          = "~"
	RedirectOperator       = ">"
	RedirectOperatorStdout = "1>"
)

var BuiltinCommands = map[string]bool{
	"echo": true,
	"type": true,
	"exit": true,
	"pwd":  true,
	"cd":   true,
}

var RedirectOperators = map[string]bool{
	">":  true,
	"1>": true,
}

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

func (cli *CLI) CommandExists(cmd string) bool {
	_, exist := BuiltinCommands[cmd]

	return exist
}

func (cli *CLI) printNotFound(cmd string) {
	fmt.Fprintf(cli.out, "%s: command not found\n", cmd)
}

func (cli *CLI) pathLookup(cmd string) (string, bool) {
	path, err := exec.LookPath(cmd)

	if err != nil {
		return "", false
	}

	return path, true
}

func (cli *CLI) RunCommand(out io.Writer, cmd string, args []string) error {
	output, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return err
	}

	fmt.Fprintln(out, strings.TrimSuffix(string(output), "\n"))

	return nil
}

func (cli *CLI) Run() {
	for {
		var arguments []string
		var redirectFile *os.File

		variableOutput := cli.out

		fmt.Fprint(cli.out, "$ ")

		inputLine := cli.ReadLine()

		if len(inputLine) == 0 {
			continue
		}

		// Index 1 element expected to be space (' ')
		commandLine, err := ParseToTokens(inputLine)

		if err != nil {
			continue
		}

		commandParts := ConsolidateTokens(commandLine)

		cmd := commandLine[0]
		arguments = commandParts[1:]

		if slices.Contains(commandParts, RedirectOperator) || slices.Contains(commandParts, RedirectOperatorStdout) {
			find := func(s string) bool {
				return slices.Contains([]string{RedirectOperator, RedirectOperatorStdout}, s)
			}
			partIdx := slices.IndexFunc(commandParts, find)
			lineIdx := slices.IndexFunc(commandLine, find)

			if partIdx+1 >= len(commandParts) {
				continue
			}

			outTarget := commandParts[partIdx+1]

			if partIdx > 1 {
				arguments = commandParts[1:partIdx]
			}

			file, err := os.OpenFile(outTarget, os.O_RDWR|os.O_CREATE, 0666)

			if err != nil {
				fmt.Fprintln(cli.out, err)
			}

			file.Truncate(0)
			file.Seek(0, io.SeekStart)

			variableOutput = file
			redirectFile = file
			commandLine = commandLine[:lineIdx-1]
		}

		switch cmd {
		case "exit":
			return
		case "echo":
			if len(commandLine) < 3 {
				continue
			}

			stream := commandLine[2:]

			fmt.Fprintf(variableOutput, "%s\n", strings.Join(stream, ""))
		case "pwd":
			dir, err := os.Getwd()

			if err != nil {
				fmt.Fprintln(cli.out, err)
			}

			fmt.Fprintln(variableOutput, dir)
		case "cd":
			if len(commandParts) < 2 {
				continue
			}

			path := commandParts[1]

			if path == homePathAlias {
				path = os.Getenv("HOME")
			}

			if err := os.Chdir(path); err != nil {
				fmt.Fprintf(cli.out, "cd: %s: No such file or directory\n", path)
			}
		case "type":
			if len(commandParts) < 2 {
				continue
			}

			typeCmd := strings.TrimSpace(commandParts[1])

			if _, exists := BuiltinCommands[typeCmd]; exists {
				fmt.Fprintf(variableOutput, "%s is a shell builtin\n", typeCmd)
			} else {
				if path, exist := cli.pathLookup(typeCmd); exist {
					fmt.Fprintf(variableOutput, "%s is %s\n", typeCmd, path)
				} else {
					fmt.Fprintf(variableOutput, "%s: not found\n", typeCmd)
				}
			}
		default:
			if _, exist := cli.pathLookup(cmd); exist {
				err := cli.RunCommand(variableOutput, cmd, arguments)

				if err != nil {
					fmt.Fprintln(cli.out, err)
				}
			} else {
				cli.printNotFound(cmd)
			}
		}

		if redirectFile != nil {
			redirectFile.Close()
		}
	}
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return strings.TrimSpace(cli.in.Text())
}
