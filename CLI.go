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
	RedirectOperatorStderr = "2>"
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

func (cli *CLI) RunCommand(out io.Writer, errOut io.Writer, cmd string, args []string) {
	command := exec.Command(cmd, args...)
	output := strings.Builder{}
	errorOutput := strings.Builder{}

	command.Stdout = &output
	command.Stderr = &errorOutput

	if err := command.Run(); err != nil {
		fmt.Fprintf(errOut, "%s", errorOutput.String())
	}

	fmt.Fprintf(out, output.String())
}

func (cli *CLI) Run() {
	for {
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

		if run := cli.runCommandLine(commandLine); !run {
			return
		}
	}
}

func (cli *CLI) runCommandLine(commandLine []string) bool {
	var arguments []string

	variableStdout := cli.out
	variableStderr := cli.out

	commandParts := ConsolidateTokens(commandLine)

	cmd := commandLine[0]
	arguments = commandParts[1:]

	if slices.Contains(commandParts, RedirectOperator) || slices.Contains(commandParts, RedirectOperatorStdout) || slices.Contains(commandParts, RedirectOperatorStderr) {
		find := func(s string) bool {
			return slices.Contains([]string{RedirectOperator, RedirectOperatorStdout, RedirectOperatorStderr}, s)
		}

		partIdx := slices.IndexFunc(commandParts, find)
		lineIdx := slices.IndexFunc(commandLine, find)

		if partIdx+1 >= len(commandParts) {
			return true
		}

		outTarget := commandParts[partIdx+1]

		if partIdx > 1 {
			arguments = commandParts[1:partIdx]
		}

		file, err := os.OpenFile(outTarget, os.O_RDWR|os.O_CREATE, 0666)

		if err != nil {
			fmt.Fprintln(cli.out, err)
		}

		defer file.Close()

		file.Truncate(0)
		file.Seek(0, io.SeekStart)

		if commandParts[partIdx] == RedirectOperatorStdout {
			variableStdout = file
		}

		if commandParts[partIdx] == RedirectOperatorStderr {
			variableStderr = file
		}

		commandLine = commandLine[:lineIdx-1]
	}

	switch cmd {
	case "exit":
		return false
	case "echo":
		if len(commandLine) < 3 {
			return true
		}

		stream := commandLine[2:]

		fmt.Fprintf(variableStdout, "%s\n", strings.Join(stream, ""))
	case "pwd":
		dir, err := os.Getwd()

		if err != nil {
			fmt.Fprintln(cli.out, err)
		}

		fmt.Fprintln(variableStdout, dir)
	case "cd":
		if len(commandParts) < 2 {
			return true
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
			return true
		}

		typeCmd := strings.TrimSpace(commandParts[1])

		if _, exists := BuiltinCommands[typeCmd]; exists {
			fmt.Fprintf(variableStdout, "%s is a shell builtin\n", typeCmd)
		} else {
			if path, exist := cli.pathLookup(typeCmd); exist {
				fmt.Fprintf(variableStdout, "%s is %s\n", typeCmd, path)
			} else {
				fmt.Fprintf(variableStdout, "%s: not found\n", typeCmd)
			}
		}
	default:
		if _, exist := cli.pathLookup(cmd); exist {
			cli.RunCommand(variableStdout, variableStderr, cmd, arguments)
		} else {
			cli.printNotFound(cmd)
		}
	}

	return true
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return strings.TrimSpace(cli.in.Text())
}
