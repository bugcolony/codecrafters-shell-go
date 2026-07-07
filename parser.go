package codecrafters_shell_go

import (
	"regexp"
	"slices"
	"strings"
)

const (
	TokenExpression = `(>{1,2}|\d*>{1,2})|("((?:[^"\\]|\\["$\\` + "`" + `])*)")+|('([^']*)')+|([^\s\\'"]+)| |\\.`
	// TokenExpression = `("((?:[^"\\]|\\["$\\` + "`" + `])*)")+|('([^']*)')+|([^\s\\'"]+)| |\\.`
	// TokenExpression = `("([^"]*)")+|('([^']*)')+|([^\s\\'"]+)| |\\.`
	// TokenExpression = `\\.|("([^"]*)")+|('([^'"]*)')+|([^\s\\'"]+)| `
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
			tokens = appendToken(tokenBuilder, tokens)

			tokenBuilder.Reset()

			quote := []rune(arg)[0]

			if string(quote) == "\"" && strings.Contains(arg, "\\") {
				tokens = append(tokens, escapeDoubleQuotedToken(arg))
			} else {
				tokens = append(tokens, strings.ReplaceAll(arg, string(quote), ""))
			}

			continue
		}

		// the debt is very technical in nature
		if arg == " " {
			tokens = appendToken(tokenBuilder, tokens)

			tokens = append(tokens, " ")

			tokenBuilder.Reset()
			continue
		}

		tokenBuilder.WriteString(arg)
	}

	tokens = appendToken(tokenBuilder, tokens)

	return tokens, nil
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

func ParseToArguments(input string) ([]string, error) {
	output, err := ParseToTokens(input)

	if err != nil {
		return nil, err
	}
	return ConsolidateTokens(output), nil
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

// ConsolidateTokens Function concat tokens if not separated by space
// so that the slice can be consumed by command runner.
func ConsolidateTokens(args []string) []string {
	var tokens []string
	stack := &strings.Builder{}

	for _, token := range args {
		if token == " " {
			tokens = appendToken(stack, tokens)
			stack.Reset()
			continue
		}

		stack.WriteString(token)
	}

	tokens = appendToken(stack, tokens)

	return tokens
}

func ParseFlag(line []string, flag string) string {
	idx := slices.Index(line, flag)

	if idx == -1 || idx+1 > len(line)-1 {
		return ""
	}

	return line[idx+1]
}
