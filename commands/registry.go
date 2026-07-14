package commands

type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

func (r *Registry) Register(command Command) {
	r.commands[command.Name()] = command
}

func (r *Registry) Get(name string) (Command, bool) {
	command, ok := r.commands[name]

	return command, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))

	for name := range r.commands {
		names = append(names, name)
	}

	return names
}

func DefaultRegistry(compReg CompletionReg, processReg *ProcessTable, history HistorySource, variables VariableRegistry) *Registry {
	r := NewRegistry()

	r.Register(Exit{})
	r.Register(&History{
		source: history,
	})
	r.Register(Pwd{})
	r.Register(Echo{})
	r.Register(&Declare{
		variables: variables,
	})
	r.Register(Cd{})
	r.Register(&Type{
		Commands: r,
	})
	r.Register(&Complete{
		CompleteRegistry: compReg,
	})
	r.Register(&Jobs{
		Process: processReg,
	})

	return r
}
