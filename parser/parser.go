package parser

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

const (
	TokenExpression = `(>{1,2}|\d*>{1,2})|("((?:[^"\\]|\\["$\\` + "`" + `])*)")+|('([^']*)')+|([^\s\\'"]+)| |\\.`
	BackgroundOp    = "&"
	PipeOp          = "|"

	RedirectOperator       = ">"
	RedirectAppend         = ">>"
	RedirectOperatorStdout = "1>"
	RedirectOperatorStderr = "2>"
	RedirectAppendStdout   = "1>>"
	RedirectAppendStderr   = "2>>"

	ErrInvalidRedirect = "invalid redirect"
)

func ParseToTokens(input string) ([]string, error) {
	var tokens []string

	argComp, err := splitToTokens(input)

	if err != nil {
		return nil, err
	}

	argComp = compactValue(argComp, " ")

	tokens = make([]string, 0, len(argComp))

	tokenBuilder := &strings.Builder{}

	for _, arg := range argComp {
		if strings.HasPrefix(arg, "\\") {
			escaped := strings.TrimPrefix(arg, "\\")

			tokenBuilder.WriteString(escaped)

			continue
		}

		if strings.HasPrefix(arg, "\"") || strings.HasPrefix(arg, "'") {
			quote := []rune(arg)[0]

			if string(quote) == "\"" && strings.Contains(arg, "\\") {
				tokenBuilder.WriteString(escapeDoubleQuotedToken(arg))
			} else {
				tokenBuilder.WriteString(strings.ReplaceAll(arg, string(quote), ""))
			}

			continue
		}

		if arg == " " {
			tokens = appendToken(tokenBuilder, tokens)

			tokenBuilder.Reset()
			continue
		}

		tokenBuilder.WriteString(arg)
	}

	tokens = appendToken(tokenBuilder, tokens)

	return tokens, nil
}

func Parse(input string) (*CommandLine, error) {
	tokens, err := ParseToTokens(input)

	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	background := false

	if tokens[len(tokens)-1] == BackgroundOp {
		tokens = tokens[:len(tokens)-1]
		background = true
	}

	if slices.Contains(tokens, PipeOp) {
		var pipeTokens [][]string
		var pipeLine []*CommandLine

		cl := &CommandLine{
			Background: background,
		}

		line := tokens

		for slices.Contains(line, PipeOp) {
			idx := slices.Index(line, PipeOp)

			if idx == -1 {
				break
			}

			pipeTokens = append(pipeTokens, line[0:idx])
			line = line[idx+1:]
		}

		if len(line) != 0 {
			pipeTokens = append(pipeTokens, line)
		}

		for _, pipeTok := range pipeTokens {
			p, err := newCommandLine(pipeTok)

			if err != nil {
				return nil, err
			}

			pipeLine = append(pipeLine, p)
		}

		if len(pipeLine) != 0 {
			cl.Pipeline = pipeLine
		}

		return cl, nil
	}

	cl, err := newCommandLine(tokens)

	if err != nil {
		return nil, err
	}

	cl.Background = background

	return cl, nil
}

func newCommandLine(tokens []string) (*CommandLine, error) {
	cl := &CommandLine{
		Name: tokens[0],
	}

	args := tokens[1:]

	findRedirectOp := func(s string) bool {
		return slices.Contains([]string{RedirectOperator, RedirectOperatorStdout, RedirectAppend, RedirectAppendStdout, RedirectOperatorStderr, RedirectAppendStderr}, s)
	}

	if slices.ContainsFunc(tokens, findRedirectOp) {
		redirect := &Redirect{}
		idx := slices.IndexFunc(tokens, findRedirectOp)

		redirectOp := tokens[idx]

		if idx+1 >= len(tokens) {
			return nil, errors.New(ErrInvalidRedirect)
		}

		redirect.Target = tokens[idx+1]

		if idx > 1 {
			args = tokens[1:idx]
		} else {
			args = []string{}
		}

		if strings.Contains(redirectOp, RedirectAppend) {
			redirect.IsAppend = true
		}

		if slices.Contains([]string{
			RedirectOperator,
			RedirectOperatorStdout,
			RedirectAppend,
			RedirectAppendStdout,
		}, redirectOp) {
			redirect.Stream = Stdout
		}

		if redirectOp == RedirectOperatorStderr || redirectOp == RedirectAppendStderr {
			redirect.Stream = Stderr
		}

		cl.Redirect = redirect
	}

	cl.Args = args

	return cl, nil
}

func appendToken(tokenBuilder *strings.Builder, tokens []string) []string {
	if tokenBuilder.Len() > 0 {
		tokens = append(tokens, tokenBuilder.String())
	}

	return tokens
}

func escapeDoubleQuotedToken(token string) string {
	stack := strings.Split(strings.Trim(token, "\""), "")
	escaped := strings.Builder{}
	escapeNext := false

	for _, char := range stack {
		if escapeNext {
			escaped.WriteString(char)
			escapeNext = false
			continue
		}

		if char == "\\" {
			escapeNext = true
			continue
		}

		escaped.WriteString(char)
	}

	return escaped.String()
}

func splitToTokens(input string) ([]string, error) {
	reg, err := regexp.Compile(TokenExpression)

	if err != nil {
		return nil, err
	}

	return reg.FindAllString(input, -1), nil
}

func compactValue(input []string, value string) []string {
	return slices.CompactFunc(input, func(a, b string) bool {
		return a == b && a == value
	})
}
