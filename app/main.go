package main

import (
	"fmt"
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

	completionReg := completion.NewRegistry()
	processReg := commands.NewProcessTable()

	registry := commands.DefaultRegistry(completionReg, processReg)
	completer := completion.NewVerboseCompleter(errOut, registry, completionReg)
	executor := shell.NewExecutor(registry, processReg)

	cli := codecraftersshellgo.NewCLI(in, out, errOut, executor, completer)

	cli.Run()
}
