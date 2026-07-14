package shell

type Variables struct {
	list map[string]string
}

func NewVariableRegistry() *Variables {
	return &Variables{
		list: make(map[string]string),
	}
}

func (v *Variables) Set(name, value string) {
	v.list[name] = value
}

func (v *Variables) Get(name string) (string, bool) {
	get, ok := v.list[name]

	return get, ok
}
