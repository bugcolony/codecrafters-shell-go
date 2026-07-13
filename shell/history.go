package shell

import (
	"bufio"
	"os"
)

type History struct {
	file *os.File
	mem  []string
}

func NewHistory(filename string) (*History, func(), error) {
	h := &History{}

	//file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)

	//if err != nil {
	//	return nil, nil, err
	//}

	cleanup := func() {
		//file.Close()
		//os.Remove(file.Name())
	}

	//h.file = file

	//h.initMem()

	return h, cleanup, nil
}

func (h *History) initMem() {
	s := bufio.NewScanner(h.file)

	for s.Scan() {
		h.mem = append(h.mem, s.Text())
	}
}

func (h *History) Push(line string) {
	h.mem = append(h.mem, line)

	//h.file.WriteString(line + "\n")
}

func (h *History) List() []string {
	return h.mem
}
