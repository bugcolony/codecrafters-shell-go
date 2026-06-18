package codecrafters_shell_go

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Run("it starts with printing $", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := strings.NewReader("exit\n")
		cli := NewCLI(in, out)

		cli.Run()

		want := "$ "
		got := out.String()

		if got != want {
			t.Errorf("expected %q to be prefixed with %q", got, want)
		}
	})

	t.Run("it rejects invalid command", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := strings.NewReader("xyz\nexit\n")
		cli := NewCLI(in, out)
		want := "$ xyz: command not found\n$ "

		cli.Run()

		if out.String() != want {
			t.Errorf("expected %s, got %s", want, out.String())
		}
	})
}
