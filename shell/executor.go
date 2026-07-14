package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/commands"
	"github.com/codecrafters-io/shell-starter-go/parser"
)

const VariableRegex = "\\$\\{([A-Za-z_][A-Za-z0-9_]*)\\}|\\$([A-Za-z_][A-Za-z0-9_]*)"

type Executor struct {
	Commands  *commands.Registry
	Processes *commands.ProcessTable
	Variables *Variables
}

func NewExecutor(commands *commands.Registry, processes *commands.ProcessTable, variables *Variables) *Executor {
	return &Executor{Commands: commands, Processes: processes, Variables: variables}
}

func (e *Executor) Execute(cl *parser.CommandLine, out io.Writer, errOut io.Writer) bool {
	defer e.Processes.ReportDone(out)

	if cl.Pipeline != nil && len(cl.Pipeline) > 0 {
		return e.runPipeLine(cl, out, errOut)
	}

	return e.runCommand(cl, nil, out, errOut)
}

func (e *Executor) runCommand(cl *parser.CommandLine, in io.Reader, out io.Writer, errOut io.Writer) bool {
	e.expand(cl)

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

func (e *Executor) expand(cl *parser.CommandLine) {
	var expanded []string
	reg, err := regexp.Compile(VariableRegex)
	replacer := strings.NewReplacer("$", "", "{", "", "}", "")

	if err != nil {
		return
	}

	for _, arg := range cl.Args {
		if strings.Contains(arg, "$") {
			expandedArgument := arg

			varCandidates := reg.FindAllString(arg, -1)

			for _, varCandidate := range varCandidates {

				varName := replacer.Replace(varCandidate)

				v, ok := e.Variables.Get(varName)

				if ok {
					expandedArgument = strings.ReplaceAll(expandedArgument, varCandidate, v)
				} else {
					expandedArgument = strings.ReplaceAll(expandedArgument, varCandidate, "")
				}
			}

			if len(expandedArgument) > 0 {
				expanded = append(expanded, expandedArgument)
			}

			continue
		}

		expanded = append(expanded, arg)
	}

	cl.Args = expanded
}
