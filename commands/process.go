package commands

type Process struct {
	Pid     int
	Command string
	State   string
}

type ProcessList struct {
	processes map[int]*Process
}

func NewProcessList() *ProcessList {
	return &ProcessList{processes: make(map[int]*Process)}
}

func (p ProcessList) List() map[int]*Process {
	return p.processes
}

func (p ProcessList) AddNewProcess(pid int, cmd string) {
	p.processes[pid] = &Process{Pid: pid, Command: cmd, State: "Running"}
}
