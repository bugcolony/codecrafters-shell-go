package codecrafters_shell_go

import (
	"regexp"
	"slices"
	"strings"
)

const (
	TokenExpression = `\\.|("([^"]*)")+|('([^'"]*)')+|([^\s\\'"]+)| `
)

func ParseToTokens(input []string) ([]string, error) {
	var tokens []string

	concat := strings.Join(input, " ")

	argComp, err := splitToTokens(concat)

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

			tokens = append(tokens, strings.ReplaceAll(arg, string(quote), ""))
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

func ParseToArguments(input []string) ([]string, error) {
	output, err := ParseToTokens(input)

	if err != nil {
		return nil, err
	}
	return consolidateTokens(output), nil
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

func consolidateTokens(args []string) []string {
	return slices.DeleteFunc(args, func(s string) bool {
		return s == " " || s == ""
	})
}
