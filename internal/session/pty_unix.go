//go:build unix

package session

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
)

type unixProc struct {
	cmd  *exec.Cmd
	file *os.File
	win  chan os.Signal
	once sync.Once
}

func startPTY(name string, args []string) (ptyProcess, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	file, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	p := &unixProc{cmd: cmd, file: file, win: make(chan os.Signal, 1)}
	if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
		_ = pty.Setsize(file, ws)
	} else {
		_ = pty.Setsize(file, &pty.Winsize{Rows: 24, Cols: 80})
	}
	signal.Notify(p.win, syscall.SIGWINCH)
	go func() {
		for range p.win {
			if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
				_ = pty.Setsize(file, ws)
			}
		}
	}()
	return p, nil
}

func (p *unixProc) Read(b []byte) (int, error)  { return p.file.Read(b) }
func (p *unixProc) Write(b []byte) (int, error) { return p.file.Write(b) }

func (p *unixProc) Close() error {
	p.once.Do(func() {
		signal.Stop(p.win)
		close(p.win)
	})
	return p.file.Close()
}

func (p *unixProc) Wait() (int, error) {
	err := p.cmd.Wait()
	if err == nil {
		return 0, nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), nil
	}
	return 1, err
}

func (p *unixProc) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func (p *unixProc) Resize(rows, cols uint16) error {
	if p.file == nil {
		return fmt.Errorf("pty closed")
	}
	return pty.Setsize(p.file, &pty.Winsize{Rows: rows, Cols: cols})
}
