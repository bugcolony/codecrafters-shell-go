package shell

import (
	"bufio"
	"io"
	"os"
)

const EnvHistoryFile = "HISTFILE"

type History struct {
	file      *os.File
	mem       []string
	appendIdx int
}

func NewHistory() (*History, func(), error) {
	h := &History{}
	filename := os.Getenv(EnvHistoryFile)

	if filename == "" {
		return h, func() {}, nil
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		file.Truncate(0)
		file.Close()

		_ = h.writeListToFile(filename, h.mem, io.SeekStart)
	}

	h.file = file

	h.initMem()

	return h, cleanup, nil
}

func (h *History) initMem() {
	s := bufio.NewScanner(h.file)

	for s.Scan() {
		h.mem = append(h.mem, s.Text())
	}
}

func (h *History) Push(lines ...string) {
	for _, line := range lines {
		h.mem = append(h.mem, line)
	}
}

func (h *History) List() []string {
	return h.mem
}

func (h *History) WriteToFile(filename string) error {
	return h.writeListToFile(filename, h.mem, io.SeekStart)
}

func (h *History) writeListToFile(filename string, list []string, whence int) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return err
	}

	file.Seek(0, whence)

	for _, line := range list {
		file.WriteString(line + "\n")
	}

	file.Close()

	return nil
}

func (h *History) AppendToFile(filename string) error {
	if h.appendIdx >= len(h.mem) {
		return nil
	}

	list := h.mem[h.appendIdx:]

	h.appendIdx = len(h.mem)

	return h.writeListToFile(filename, list, io.SeekEnd)
}
