package main

import (
	"fmt"
	"os"

	codecrafters_shell_go "github.com/codecrafters-io/shell-starter-go"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	out := os.Stdout
	in := os.Stdin
	cli := codecrafters_shell_go.NewCLI(in, out)

	cli.Run()
}
