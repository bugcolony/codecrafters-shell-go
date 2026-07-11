package commands

import (
	"fmt"
	"io"
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
	processes map[int]*Process
	mu        sync.Mutex
	CurrentId int
}

func NewProcessTable() *ProcessTable {
	return &ProcessTable{processes: make(map[int]*Process)}
}

func (p *ProcessTable) listLocked() []*Process {
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

func (p *ProcessTable) List() []Process {
	p.mu.Lock()
	defer p.mu.Unlock()

	var list []Process
	for _, proc := range p.listLocked() {
		list = append(list, *proc)
	}

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

func (p *ProcessTable) ReportDone(out io.Writer) {
	p.mu.Lock()
	list := p.listLocked()
	p.mu.Unlock()

	// TODO: dedupe jobs
	indicator := " "

	for idx, proc := range list {
		indicator = " "

		switch idx {
		case len(list) - 1:
			indicator = "+"
		case len(list) - 2:
			indicator = "-"
		default:
			indicator = " "
		}

		if proc.State == "Done" {
			fmt.Fprintf(out, "[%d]%s %-24s %s\n", proc.Id, indicator, proc.State, proc.Command)
		}
	}

	return
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
	idx := 1

	for idx <= p.CurrentId {
		if _, exists := p.processes[idx]; !exists {
			return idx
		}
		idx++
	}

	idx = p.CurrentId + 1

	p.CurrentId = idx

	return idx
}
