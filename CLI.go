package codecrafters_shell_go

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	homePathAlias = "~"
)

var BuiltinCommands = map[string]bool{
	"echo": true,
	"type": true,
	"exit": true,
	"pwd":  true,
	"cd":   true,
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
	// or just use exec.LookPath(cmd)
	path := os.Getenv("PATH")

	dirs := strings.Split(path, string(os.PathListSeparator))

	for _, dir := range dirs {
		// check filesystem file exists and has x perm
		fs, _ := os.ReadDir(dir)
		for _, f := range fs {
			fileInfo, err := f.Info()

			if err != nil {
				continue
			}

			if f.Name() == cmd && fileInfo.Mode()&0111 != 0 {
				return filepath.Join(dir, f.Name()), true
			}
		}
	}

	return "", false
}

func (cli *CLI) RunCommand(cmd string, args []string) error {
	output, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return err
	}

	fmt.Fprintln(cli.out, strings.TrimSuffix(string(output), "\n"))

	return nil
}

func (cli *CLI) sanitizeArguments(raw []string) ([]string, error) {
	tokens, err := ParseToTokens(raw)

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (cli *CLI) Run() {
	for {
		fmt.Fprint(cli.out, "$ ")

		inputLine := strings.Split(cli.ReadLine(), " ")

		if len(inputLine) == 0 {
			continue
		}

		commandLine, err := ParseToArguments(inputLine)

		if err != nil {
			continue
		}

		cmd := commandLine[0]

		switch cmd {
		case "exit":
			return
		case "echo":
			args, err := cli.sanitizeArguments(inputLine[1:])
			if err != nil {
				fmt.Fprintln(cli.out, err)
				continue
			}

			fmt.Fprintf(cli.out, "%s\n", strings.Join(args, ""))
		case "pwd":
			dir, err := os.Getwd()

			if err != nil {
				fmt.Fprintln(cli.out, err)
			}

			fmt.Fprintln(cli.out, dir)
		case "cd":
			if len(inputLine) < 2 {
				continue
			}

			path := inputLine[1]

			if path == homePathAlias {
				path = os.Getenv("HOME")
			}

			if err := os.Chdir(path); err != nil {
				fmt.Fprintf(cli.out, "cd: %s: No such file or directory\n", path)
			}
		case "type":
			if len(inputLine) < 2 {
				continue
			}

			arg := strings.TrimSpace(inputLine[1])

			if _, exists := BuiltinCommands[arg]; exists {
				fmt.Fprintf(cli.out, "%s is a shell builtin\n", arg)
			} else {
				if path, exist := cli.pathLookup(arg); exist {
					fmt.Fprintf(cli.out, "%s is %s\n", arg, path)
				} else {
					fmt.Fprintf(cli.out, "%s: not found\n", arg)
				}
			}
		default:
			extCmd := strings.TrimSpace(cmd)

			if _, exist := cli.pathLookup(extCmd); exist {
				var arguments []string

				if len(inputLine) > 1 {
					arguments, _ = ParseToArguments(inputLine[1:])
				}

				err := cli.RunCommand(extCmd, arguments)
				if err != nil {
					fmt.Fprintln(cli.out, err)
				}
			} else {
				cli.printNotFound(cmd)
			}
		}
	}
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return cli.in.Text()
}
