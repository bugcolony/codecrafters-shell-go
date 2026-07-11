package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/commands"
	"github.com/codecrafters-io/shell-starter-go/parser"
)

type Executor struct {
	Commands  *commands.Registry
	Processes *commands.ProcessList
}

func NewExecutor(commands *commands.Registry, processes *commands.ProcessList) *Executor {
	return &Executor{Commands: commands, Processes: processes}
}

func (e *Executor) Execute(cl *parser.CommandLine, out io.Writer, errOut io.Writer) bool {
	name := cl.Name
	args := cl.Args
	stdout := out
	stderr := errOut
	cleanup := func() {}

	if cl.Redirect != nil {
		redirectOut, redirectErr, cleanupFunc, ok := applyRedirect(cl.Redirect, out, errOut)

		if !ok {
			return true
		}

		stdout = redirectOut
		stderr = redirectErr
		cleanup = cleanupFunc
	}

	defer cleanup()

	if cmd, exists := e.Commands.Get(name); exists {
		return cmd.Execute(args, stdout, stderr)
	}

	return e.executeExternal(cl, stdout, stderr)
}

func (e *Executor) executeExternal(cl *parser.CommandLine, out io.Writer, errOut io.Writer) bool {
	_, err := exec.LookPath(cl.Name)

	if err != nil {
		fmt.Fprintf(errOut, "%s: command not found\n", cl.Name)

		return true
	}

	command := exec.Command(cl.Name, cl.Args...)

	command.Stdout = out
	command.Stderr = errOut

	if cl.Background {
		err := command.Start()
		if err != nil {
			fmt.Fprintf(errOut, "%s: command could not start\n", cl.Name)

			return true
		}

		pid := command.Process.Pid

		e.Processes.AddNewProcess(pid, strings.TrimSpace(fmt.Sprintf("%s %s", cl.Name, strings.Join(cl.Args, " "))))

		fmt.Fprintf(out, "[1] %d\n", pid)

		go func() {
			err := command.Wait()

			if err != nil {
				return
			}
		}()
	} else {
		_ = command.Run()
	}

	return true
}

func applyRedirect(r *parser.Redirect, out io.Writer, errOut io.Writer) (io.Writer, io.Writer, func(), bool) {
	variableOut := out
	variableErrOut := errOut

	file, err := os.OpenFile(r.Target, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		fmt.Fprintln(out, err)

		return nil, nil, nil, false
	}

	cleanup := func() {
		file.Close()
	}

	if r.IsAppend {
		file.Seek(0, io.SeekEnd)
	} else {
		file.Truncate(0)
		file.Seek(0, io.SeekStart)
	}

	if r.Stream == parser.Stdout {
		variableOut = file
	}

	if r.Stream == parser.Stderr {
		variableErrOut = file
	}

	return variableOut, variableErrOut, cleanup, true
}
