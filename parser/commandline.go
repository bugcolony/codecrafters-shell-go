package parser

type RedirectStream int

const (
	Stdin RedirectStream = iota
	Stdout
	Stderr
)

type CommandLine struct {
	Name       string
	Args       []string
	Redirect   *Redirect
	Background bool
	Pipeline   []*CommandLine
}

type Redirect struct {
	Stream   RedirectStream
	Target   string
	IsAppend bool
}
