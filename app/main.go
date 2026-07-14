package main

import (
	"fmt"
	"log"
	"os"

	codecraftersshellgo "github.com/codecrafters-io/shell-starter-go"
	"github.com/codecrafters-io/shell-starter-go/commands"
	"github.com/codecrafters-io/shell-starter-go/completion"
	"github.com/codecrafters-io/shell-starter-go/shell"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	out := os.Stdout
	in := os.Stdin
	errOut := os.Stderr

	history, cleanup, err := shell.NewHistory()

	if err != nil {
		log.Fatal(err)
	}

	defer cleanup()

	completionReg := completion.NewRegistry()
	processReg := commands.NewProcessTable()
	variableReg := shell.NewVariableRegistry()

	registry := commands.DefaultRegistry(completionReg, processReg, history, variableReg)
	completer := completion.NewVerboseCompleter(errOut, registry, completionReg)
	executor := shell.NewExecutor(registry, processReg, variableReg)

	cli := codecraftersshellgo.NewCLI(in, out, errOut, executor, completer, history)

	cli.Run()
}
