package codecrafters_shell_go

import (
	"bufio"
	"io"

	"github.com/chzyer/readline"
	"github.com/codecrafters-io/shell-starter-go/completion"
	"github.com/codecrafters-io/shell-starter-go/parser"
	"github.com/codecrafters-io/shell-starter-go/shell"
)

const (
	Prompt = "$ "
)

type CLI struct {
	in        *bufio.Scanner
	out       io.Writer
	errOut    io.Writer
	executor  *shell.Executor
	completer *completion.VerboseCompleter
	history   *shell.History
}

func NewCLI(in io.Reader, out io.Writer, errOut io.Writer, executor *shell.Executor, completer *completion.VerboseCompleter, history *shell.History) *CLI {
	return &CLI{
		in:        bufio.NewScanner(in),
		out:       out,
		errOut:    errOut,
		executor:  executor,
		completer: completer,
		history:   history,
	}
}

func (cli *CLI) Run() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          Prompt,
		AutoComplete:    cli.completer,
		InterruptPrompt: "^C",
		Stdout:          cli.out,
		Stderr:          cli.errOut,
	})

	if err != nil {
		panic(err)
	}

	cli.completer.AttachReadline(rl)

	defer rl.Close()

	for {
		inputLine, err := rl.Readline()

		if err != nil {
			break
		}

		if len(inputLine) == 0 {
			continue
		}

		cli.history.Push(inputLine)

		commandLine, err := parser.Parse(inputLine)

		if err != nil || commandLine == nil {
			continue
		}

		if run := cli.executor.Execute(commandLine, cli.out, cli.errOut); !run {
			return
		}
	}
}
