package completion

type Registry struct {
	Scripts map[string]string
}

func NewRegistry() *Registry {
	return &Registry{Scripts: make(map[string]string)}
}

func (r *Registry) Set(name, path string) {
	r.Scripts[name] = path
}

func (r *Registry) Get(name string) (string, bool) {
	comp, ok := r.Scripts[name]

	return comp, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.Scripts))

	for name := range r.Scripts {
		names = append(names, name)
	}

	return names
}
