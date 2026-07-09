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

	registry := commands.DefaultRegistry(completionReg)
	completer := completion.NewVerboseCompleter(errOut, registry, completionReg)
	executor := shell.NewExecutor(registry)

	cli := codecraftersshellgo.NewCLI(in, out, errOut, executor, completer)

	cli.Run()
}
