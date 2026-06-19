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
				if path, exist := cli.pathLookup(arg); exist {
					fmt.Fprintf(cli.out, "%s is %s\n", arg, path)
				} else {
					fmt.Fprintf(cli.out, "%s: not found\n", arg)
				}
			}
		default:
			extCmd := strings.TrimSpace(inputLine[0])

			if _, exist := cli.pathLookup(extCmd); exist {
				var arguments []string

				if len(inputLine) > 1 {
					arguments = inputLine[1:]
				}

				cli.RunCommand(extCmd, arguments)
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
