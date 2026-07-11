package commands

import (
	"maps"
	"slices"
	"sync"
)

type Process struct {
	Id      int
	Pid     int
	Command string
	State   string
}

type ProcessTable struct {
	processes       map[int]*Process
	mu              sync.Mutex
	CurrentProcess  int
	PreviousProcess int
}

func NewProcessTable() *ProcessTable {
	return &ProcessTable{processes: make(map[int]*Process)}
}

func (p *ProcessTable) List() []*Process {
	p.mu.Lock()
	defer p.mu.Unlock()

	var list []*Process

	for _, key := range slices.Sorted(maps.Keys(p.processes)) {

		p, found := p.processes[key]

		if found {
			list = append(list, p)
		}
	}

	p.purge()

	return list
}

func (p *ProcessTable) purge() {
	for _, proc := range p.processes {
		if proc.State == "Done" {
			delete(p.processes, proc.Id)
		}
	}
}

func (p *ProcessTable) AddNewProcess(pid int, cmd string) *Process {
	p.mu.Lock()
	defer p.mu.Unlock()

	proc := &Process{
		Id:      p.generateId(),
		Pid:     pid,
		Command: cmd,
		State:   "Running",
	}

	p.processes[proc.Id] = proc

	return proc
}

func (p *ProcessTable) MarkDone(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, ok := p.processes[id]

	if ok {
		f.State = "Done"
	}
}

func (p *ProcessTable) generateId() int {
	idx := p.CurrentProcess + 1

	p.PreviousProcess = p.CurrentProcess

	p.CurrentProcess = idx

	return idx
}
