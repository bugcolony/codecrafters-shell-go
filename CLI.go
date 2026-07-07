package codecrafters_shell_go

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/chzyer/readline"
)

const (
	HomePathAlias = "~"
	Prompt        = "$ "

	RedirectOperator       = ">"
	RedirectAppend         = ">>"
	RedirectOperatorStdout = "1>"
	RedirectOperatorStderr = "2>"
	RedirectAppendStdout   = "1>>"
	RedirectAppendStderr   = "2>>"
)

var BuiltinCommands = map[string]bool{
	"echo":     true,
	"type":     true,
	"exit":     true,
	"pwd":      true,
	"cd":       true,
	"complete": true,
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("exit"),
	readline.PcItem("complete"),
	readline.PcItemDynamic(searchPath(),
		readline.PcItemDynamic(searchFile()),
	),
)

func searchPath() func(string) []string {
	return func(line string) []string {
		if line == "" {
			return nil
		}
		command, _, _ := strings.Cut(line, " ")

		var candidates []string
		var wg sync.WaitGroup
		var mu sync.Mutex

		envPath := os.Getenv("PATH")

		dirs := strings.Split(envPath, string(os.PathListSeparator))

		for _, dir := range dirs {
			fs, err := os.ReadDir(dir)

			if err != nil {
				continue
			}

			wg.Go(func() {
				var dirCandidates []string

				for _, f := range fs {
					if strings.HasPrefix(f.Name(), command) && !f.IsDir() {
						fileInfo, err := f.Info()

						if err != nil {
							continue
						}

						if fileInfo.Mode()&0111 != 0 {
							dirCandidates = append(dirCandidates, f.Name())
						}
					}
				}

				if len(dirCandidates) > 0 {
					mu.Lock()
					candidates = append(candidates, dirCandidates...)
					mu.Unlock()
				}
			})

		}

		wg.Wait()

		slices.Sort(candidates)

		return slices.Compact(candidates)
	}
}

func searchFile() func(string) []string {
	return func(line string) []string {
		var result []string
		nested := false
		filePath := "./"
		tokens := strings.Split(line, " ")

		if len(tokens) < 2 {
			return nil
		}

		argLine := tokens[1 : len(tokens)-1]
		filename := tokens[len(tokens)-1]

		if strings.Contains(filename, string(os.PathSeparator)) {
			nested = true
			filePath = path.Clean(filePath + path.Dir(filename))

			if strings.HasSuffix(filename, string(os.PathSeparator)) {
				filename = ""
			} else {
				filename = path.Base(filename)
			}
		}

		dir, err := os.ReadDir(filePath)

		if err != nil {
			return nil
		}

		for _, f := range dir {
			if strings.HasPrefix(f.Name(), filename) {
				var name string

				if nested {
					name = path.Join(filePath, f.Name())
				} else {
					name = f.Name()
				}

				if f.IsDir() {
					name += "/"
				}

				result = append(result, strings.TrimSpace(fmt.Sprintf("%s %s", strings.Join(argLine, " "), name)))
			}
		}

		return result
	}
}

type verboseCompleter struct {
	inner    readline.AutoCompleter
	readline *readline.Instance
	stderr   io.Writer
	lastLine []rune
}

func (v *verboseCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if len(line) == 0 {
		return nil, 0
	}

	var lineSuggestions [][]rune
	newLine, offset := v.inner.Do(line, pos)

	for _, s := range newLine {
		candidate := s

		if strings.HasSuffix(string(s), "/ ") {
			candidate = []rune(strings.TrimSuffix(string(s), " "))
		}

		lineSuggestions = append(lineSuggestions, candidate)
	}

	if len(newLine) == 0 {
		fmt.Fprint(v.readline.Stderr(), "\a")
	}

	if len(newLine) > 1 {
		var suggestions []string
		input := string(line)
		autoCompPrefix := input

		tokens := strings.SplitAfter(input, " ")

		if len(tokens) > 1 {
			autoCompPrefix = tokens[len(tokens)-1]
		}

		for _, line := range newLine {
			suggestions = append(suggestions, autoCompPrefix+string(line))
		}

		autoFill := longestCommonPrefixAutoFill(autoCompPrefix, suggestions)

		if autoFill != "" {
			v.readline.Operation.SetBuffer(input + autoFill)
			v.lastLine = []rune(input + autoFill)

			return nil, 0
		}

		if !slices.Equal(line, v.lastLine) {
			fmt.Fprint(v.readline.Stderr(), "\a")
			v.lastLine = line

			return nil, 0
		}

		v.readline.Terminal.Write([]byte(fmt.Sprintln("\n" + strings.Join(suggestions, "  "))))

		v.readline.Operation.SetBuffer(string(line))

		return nil, offset
	}

	v.lastLine = line

	return lineSuggestions, offset
}

func longestCommonPrefixAutoFill(prefix string, suggestions []string) string {
	result := ""
	candidate := slices.MinFunc(suggestions, func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})

	candidate = strings.TrimPrefix(candidate, prefix)

	for _, char := range strings.Split(candidate, "") {
		if matchAllPrefix(suggestions, prefix+result+char) {
			result += char
		} else {
			break
		}
	}

	return result
}

func matchAllPrefix(suggestions []string, prefix string) bool {
	matchedCases := 0

	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, prefix) {
			matchedCases++
		}
	}

	return matchedCases == len(suggestions)
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
	executable, err := exec.LookPath(cmd)

	if err != nil {
		return "", false
	}

	return executable, true
}

func (cli *CLI) RunCommand(out io.Writer, errOut io.Writer, cmd string, args []string) {
	command := exec.Command(cmd, args...)
	output := strings.Builder{}
	errorOutput := strings.Builder{}

	command.Stdout = &output
	command.Stderr = &errorOutput

	if err := command.Run(); err != nil {
		fmt.Fprintf(errOut, "%s", errorOutput.String())
	}

	fmt.Fprint(out, output.String())
}

func (cli *CLI) Run() {

	vc := &verboseCompleter{inner: completer, stderr: os.Stderr}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          Prompt,
		AutoComplete:    vc,
		InterruptPrompt: "^C",
		Stdout:          cli.out,
		Stderr:          os.Stderr,
	})

	if err != nil {
		panic(err)
	}

	vc.readline = rl

	defer rl.Close()

	for {
		inputLine, err := rl.Readline()

		if err != nil {
			break
		}

		if len(inputLine) == 0 {
			continue
		}

		commandLine, err := ParseToTokens(inputLine)

		if err != nil {
			continue
		}

		if run := cli.runCommandLine(commandLine); !run {
			return
		}
	}
}

func (cli *CLI) runCommandLine(commandLine []string) bool {
	var arguments []string

	variableStdout := cli.out
	variableStderr := cli.out

	commandParts := ConsolidateTokens(commandLine)

	cmd := commandLine[0]
	arguments = commandParts[1:]
	findRedirectOp := func(s string) bool {
		return slices.Contains([]string{RedirectOperator, RedirectOperatorStdout, RedirectAppend, RedirectAppendStdout, RedirectOperatorStderr, RedirectAppendStderr}, s)
	}

	if slices.ContainsFunc(commandParts, findRedirectOp) {
		partIdx := slices.IndexFunc(commandParts, findRedirectOp)
		lineIdx := slices.IndexFunc(commandLine, findRedirectOp)
		redirectOp := commandParts[partIdx]

		if partIdx+1 >= len(commandParts) {
			return true
		}

		outTarget := commandParts[partIdx+1]

		if partIdx > 1 {
			arguments = commandParts[1:partIdx]
		} else {
			arguments = []string{}
		}

		file, err := os.OpenFile(outTarget, os.O_RDWR|os.O_CREATE, 0666)

		if err != nil {
			fmt.Fprintln(cli.out, err)
		}

		defer file.Close()

		if strings.Contains(redirectOp, RedirectAppend) {
			file.Seek(0, io.SeekEnd)
		} else {
			file.Truncate(0)
			file.Seek(0, io.SeekStart)
		}

		if slices.Contains([]string{
			RedirectOperator,
			RedirectOperatorStdout,
			RedirectAppend,
			RedirectAppendStdout,
		}, redirectOp) {
			variableStdout = file
		}

		if redirectOp == RedirectOperatorStderr || redirectOp == RedirectAppendStderr {
			variableStderr = file
		}

		commandLine = commandLine[:lineIdx-1]
	}

	switch cmd {
	case "exit":
		return false
	case "echo":
		if len(commandLine) < 3 {
			return true
		}

		stream := commandLine[2:]

		fmt.Fprintf(variableStdout, "%s\n", strings.Join(stream, ""))
	case "complete":

	case "pwd":
		dir, err := os.Getwd()

		if err != nil {
			fmt.Fprintln(cli.out, err)
		}

		fmt.Fprintln(variableStdout, dir)
	case "cd":
		if len(commandParts) < 2 {
			return true
		}

		pathArg := commandParts[1]

		if pathArg == HomePathAlias {
			pathArg = os.Getenv("HOME")
		}

		if err := os.Chdir(pathArg); err != nil {
			fmt.Fprintf(cli.out, "cd: %s: No such file or directory\n", pathArg)
		}
	case "type":
		if len(commandParts) < 2 {
			return true
		}

		typeCmd := strings.TrimSpace(commandParts[1])

		if _, exists := BuiltinCommands[typeCmd]; exists {
			fmt.Fprintf(variableStdout, "%s is a shell builtin\n", typeCmd)
		} else {
			if lookup, exist := cli.pathLookup(typeCmd); exist {
				fmt.Fprintf(variableStdout, "%s is %s\n", typeCmd, lookup)
			} else {
				fmt.Fprintf(variableStdout, "%s: not found\n", typeCmd)
			}
		}
	default:
		if _, exist := cli.pathLookup(cmd); exist {
			cli.RunCommand(variableStdout, variableStderr, cmd, arguments)
		} else {
			cli.printNotFound(cmd)
		}
	}

	return true
}

func (cli *CLI) ReadLine() string {
	cli.in.Scan()

	return strings.TrimSpace(cli.in.Text())
}
