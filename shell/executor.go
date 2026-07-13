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
	Processes *commands.ProcessTable
}

func NewExecutor(commands *commands.Registry, processes *commands.ProcessTable) *Executor {
	return &Executor{Commands: commands, Processes: processes}
}

func (e *Executor) Execute(cl *parser.CommandLine, out io.Writer, errOut io.Writer) bool {
	defer e.Processes.ReportDone(out)

	if cl.Pipeline != nil && len(cl.Pipeline) > 0 {
		return e.runPipeLine(cl, out, errOut)
	}

	return e.runCommand(cl, nil, out, errOut)
}

func (e *Executor) runCommand(cl *parser.CommandLine, in io.Reader, out io.Writer, errOut io.Writer) bool {
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

	if cmd, exists := e.Commands.Get(cl.Name); exists {
		return cmd.Execute(cl.Args, stdout, stderr)
	}

	return e.executeExternal(cl, in, stdout, stderr)
}

func (e *Executor) runPipeLine(cl *parser.CommandLine, out io.Writer, errOut io.Writer) bool {
	var lastReader io.Reader

	for idx, cmd := range cl.Pipeline {
		if idx == len(cl.Pipeline)-1 {
			e.runCommand(cmd, lastReader, out, errOut)
			continue
		}

		pRead, pWrite := io.Pipe()
		in := lastReader

		go func() {
			defer pWrite.Close()

			e.runCommand(cmd, in, pWrite, errOut)
		}()

		lastReader = pRead
	}

	return true
}

func (e *Executor) executeExternal(cl *parser.CommandLine, in io.Reader, out io.Writer, errOut io.Writer) bool {
	_, err := exec.LookPath(cl.Name)

	if err != nil {
		fmt.Fprintf(errOut, "%s: command not found\n", cl.Name)

		return true
	}

	command := exec.Command(cl.Name, cl.Args...)

	if in != nil {
		command.Stdin = in
	}

	command.Stdout = out
	command.Stderr = errOut

	if cl.Background {
		err := command.Start()
		if err != nil {
			fmt.Fprintf(errOut, "%s: command could not start\n", cl.Name)

			return true
		}

		pid := command.Process.Pid

		proc := e.Processes.AddNewProcess(pid, strings.TrimSpace(fmt.Sprintf("%s %s", cl.Name, strings.Join(cl.Args, " "))))

		fmt.Fprintf(out, "[%d] %d\n", proc.Id, pid)

		go func() {
			err := command.Wait()

			if err != nil {
				return
			}

			e.Processes.MarkDone(proc.Id)
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
