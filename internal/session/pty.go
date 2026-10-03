package session

import "io"

type ptyProcess interface {
	io.ReadWriteCloser
	Wait() (int, error)
	Kill() error
	Resize(rows, cols uint16) error
}
