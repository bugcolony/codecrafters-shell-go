package shell

import (
	"bufio"
	"os"
)

type History struct {
	file *os.File
	mem  [][]byte
}

func NewHistory(filename string) (*History, func(), error) {
	h := &History{}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		file.Close()
	}

	h.file = file

	h.initMem()

	return h, cleanup, nil
}

func (h *History) initMem() {
	s := bufio.NewScanner(h.file)

	for s.Scan() {
		h.mem = append(h.mem, s.Bytes())
	}
}

func (h *History) Push(line string) {
	h.mem = append(h.mem, []byte(line))

	h.file.WriteString(line + "\n")
}

func (h *History) List() [][]byte {
	return h.mem
}
