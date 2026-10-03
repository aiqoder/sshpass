package session

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/aiqoder/sshpass/internal/prompt"
)

type Options struct {
	Command        []string
	Password       string
	PasswordPrompt string
	HostPrompt     string
	HostConfirm    bool
	Timeout        time.Duration
	Verbose        bool
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
}

func Run(ctx context.Context, opt Options) (int, error) {
	if len(opt.Command) == 0 {
		return 1, fmt.Errorf("missing command")
	}
	if opt.Stdin == nil {
		opt.Stdin = os.Stdin
	}
	if opt.Stdout == nil {
		opt.Stdout = os.Stdout
	}
	if opt.Stderr == nil {
		opt.Stderr = os.Stderr
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 30 * time.Second
	}

	proc, err := startPTY(opt.Command[0], opt.Command[1:])
	if err != nil {
		return 127, fmt.Errorf("start command: %w", err)
	}
	defer proc.Close()

	logf := func(format string, args ...any) {
		if opt.Verbose {
			fmt.Fprintf(opt.Stderr, "sshpass: "+format+"\n", args...)
		}
	}

	if err := injectPrompts(ctx, proc, opt, logf); err != nil {
		_ = proc.Kill()
		_, _ = proc.Wait()
		return 1, err
	}

	go func() {
		_, _ = io.Copy(proc, opt.Stdin)
	}()

	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(opt.Stdout, proc)
		close(copyDone)
	}()

	code, waitErr := proc.Wait()
	select {
	case <-copyDone:
	case <-time.After(2 * time.Second):
	}
	if waitErr != nil {
		return code, waitErr
	}
	return code, nil
}

func injectPrompts(ctx context.Context, proc ptyProcess, opt Options, logf func(string, ...any)) error {
	ctx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	matcher := prompt.New(opt.PasswordPrompt, opt.HostPrompt, opt.HostConfirm)
	var (
		done        atomic.Bool
		hostAnswers int
		buf         []byte
	)
	const maxHostAnswers = 8

	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded && !done.Load() {
			logf("timed out waiting for password prompt")
			_ = proc.Kill()
		}
	}()

	tmp := make([]byte, 4096)
	for {
		if ctx.Err() != nil && !done.Load() {
			return fmt.Errorf("timed out waiting for password prompt")
		}
		n, err := proc.Read(tmp)
		if n > 0 {
			chunk := tmp[:n]
			if _, werr := opt.Stdout.Write(chunk); werr != nil {
				return werr
			}
			buf = append(buf, chunk...)
			for hostAnswers < maxHostAnswers {
				matched, rest := matcher.HostMatched(buf)
				if !matched {
					break
				}
				logf("answered host key confirmation")
				if _, werr := proc.Write([]byte("yes\n")); werr != nil {
					return fmt.Errorf("write host confirmation: %w", werr)
				}
				hostAnswers++
				buf = rest
			}
			if matched, _ := matcher.PasswordMatched(buf); matched {
				logf("sending password")
				if _, werr := proc.Write([]byte(opt.Password + "\n")); werr != nil {
					return fmt.Errorf("write password: %w", werr)
				}
				done.Store(true)
				return nil
			}
			buf = prompt.TrimTail(buf, 8192)
		}
		if err != nil {
			if done.Load() {
				return nil
			}
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("timed out waiting for password prompt")
			}
			if err == io.EOF {
				return fmt.Errorf("command exited before password prompt")
			}
			return fmt.Errorf("read pty: %w", err)
		}
	}
}
