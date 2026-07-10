package completion

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/chzyer/readline"
)

const (
	EnvCompLineKey  = "COMP_LINE"
	EnvCompPointKey = "COMP_POINT"
)

type CommandLister interface {
	Names() []string
}

type VerboseCompleter struct {
	CommandList        CommandLister
	CompletionRegistry *Registry

	inner    readline.AutoCompleter
	readline *readline.Instance
	stderr   io.Writer
	lastLine []rune
}

func NewVerboseCompleter(stderr io.Writer, commandList CommandLister, completionRegistry *Registry) *VerboseCompleter {
	v := &VerboseCompleter{
		stderr:             stderr,
		CommandList:        commandList,
		CompletionRegistry: completionRegistry,
	}

	v.inner = readline.NewPrefixCompleter(
		readline.PcItem("exit"),
		readline.PcItem("complete",
			readline.PcItem("-p"),
			readline.PcItem("-C"),
		),
		readline.PcItemDynamic(v.PathCandidates,
			readline.PcItemDynamic(v.FileCandidates),
		),
	)

	return v
}

func (v *VerboseCompleter) AttachReadline(readline *readline.Instance) {
	v.readline = readline
}

func (v *VerboseCompleter) Do(line []rune, pos int) ([][]rune, int) {
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

		autoFill := v.longestCommonPrefixAutoFill(autoCompPrefix, suggestions)

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

func (v *VerboseCompleter) longestCommonPrefixAutoFill(prefix string, suggestions []string) string {
	result := ""
	candidate := slices.MinFunc(suggestions, func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})

	candidate = strings.TrimPrefix(candidate, prefix)

	for _, char := range strings.Split(candidate, "") {
		if v.matchAllPrefix(suggestions, prefix+result+char) {
			result += char
		} else {
			break
		}
	}

	return result
}

func (v *VerboseCompleter) matchAllPrefix(suggestions []string, prefix string) bool {
	matchedCases := 0

	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, prefix) {
			matchedCases++
		}
	}

	return matchedCases == len(suggestions)
}

func (v *VerboseCompleter) PathCandidates(line string) []string {
	if line == "" {
		return nil
	}

	command, _, _ := strings.Cut(line, " ")

	var candidates []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	candidates = append(candidates, v.CompletionRegistry.Names()...)

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

func (v *VerboseCompleter) FileCandidates(line string) []string {
	var result []string
	var argLine []string
	filename := ""
	nested := false
	filePath := "./"
	tokens := strings.Split(strings.TrimSpace(line), " ")

	if len(tokens) == 0 {
		return nil
	}

	if f, ok := v.CompletionRegistry.Get(tokens[0]); ok {
		return v.completionRegistryCandidates(f, line)
	}

	if len(tokens) > 1 {
		argLine = tokens[1 : len(tokens)-1]
		filename = tokens[len(tokens)-1]
	}

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

func (v *VerboseCompleter) completionRegistryCandidates(script string, line string) []string {
	tokens := strings.Split(strings.TrimSpace(line), " ")
	argCommand := tokens[0]
	argCompletion := tokens[len(tokens)-1]
	argPrev := ""
	var argLine []string

	_ = os.Setenv(EnvCompLineKey, line)
	_ = os.Setenv(EnvCompPointKey, strconv.Itoa(len(line)))

	if len(tokens) == 1 {
		argPrev = argCommand
	}

	if len(tokens) > 2 {
		argPrev = tokens[len(tokens)-2]
		argLine = tokens[1 : len(tokens)-1]
	}

	out := bytes.Buffer{}

	cmd := exec.Command(script, argCommand, argCompletion, argPrev)

	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil
	}

	var res []string

	for _, candidate := range strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n") {
		res = append(res, strings.TrimSpace(fmt.Sprintf("%s %s", strings.Join(argLine, " "), candidate)))
	}

	return res
}
