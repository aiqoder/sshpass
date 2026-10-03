package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

const (
	DefaultPasswordPrompt = "assword"
	DefaultHostPrompt     = "are you sure you want to continue connecting (yes/no"
	DefaultTimeout        = 30 * time.Second
)

type Options struct {
	Password       string
	UseEnv         bool
	File           string
	FD             int
	PasswordPrompt string
	HostPrompt     string
	HostConfirm    bool
	Timeout        time.Duration
	Verbose        bool
	Command        []string
}

func Parse(args []string, stderr io.Writer) (Options, error) {
	fs := flag.NewFlagSet("sshpass", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var opt Options
	opt.FD = -1
	var timeoutSec int
	var noHostConfirm bool
	var enableHostConfirm bool

	fs.StringVar(&opt.Password, "p", "", "password (visible in process list; prefer -e)")
	fs.BoolVar(&opt.UseEnv, "e", false, "read password from SSHPASS environment variable")
	fs.StringVar(&opt.File, "f", "", "read password from the first line of file")
	fs.IntVar(&opt.FD, "d", -1, "read password from file descriptor (Unix)")
	fs.StringVar(&opt.PasswordPrompt, "P", DefaultPasswordPrompt, "password prompt substring to match")
	fs.StringVar(&opt.HostPrompt, "H", DefaultHostPrompt, "host-key confirmation prompt substring")
	fs.BoolVar(&enableHostConfirm, "y", false, "enable host-key auto-confirm (default on)")
	fs.BoolVar(&noHostConfirm, "no-host-confirm", false, "do not auto-answer host-key confirmation")
	fs.IntVar(&timeoutSec, "t", int(DefaultTimeout.Seconds()), "seconds to wait for the password prompt")
	fs.BoolVar(&opt.Verbose, "v", false, "verbose diagnostics on stderr")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: sshpass [-p pass|-e|-f file|-d fd] [-P prompt] [-H host-prompt] [-y] [--no-host-confirm] [-t sec] [-v] -- command [args...]\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	opt.Command = fs.Args()
	if len(opt.Command) == 0 {
		fs.Usage()
		return Options{}, errors.New("missing command")
	}

	// Host confirm is on by default; --no-host-confirm disables it.
	// -y is accepted for explicit scripts and keeps confirm enabled.
	opt.HostConfirm = !noHostConfirm || enableHostConfirm
	if timeoutSec <= 0 {
		return Options{}, errors.New("timeout must be positive")
	}
	opt.Timeout = time.Duration(timeoutSec) * time.Second

	if opt.PasswordPrompt == "" {
		opt.PasswordPrompt = DefaultPasswordPrompt
	}
	if opt.HostPrompt == "" {
		opt.HostPrompt = DefaultHostPrompt
	}

	n := 0
	if opt.Password != "" {
		n++
	}
	if opt.UseEnv {
		n++
	}
	if opt.File != "" {
		n++
	}
	if opt.FD >= 0 {
		n++
	}
	if n == 0 {
		fs.Usage()
		return Options{}, errors.New("password source required: use -p, -e, -f, or -d")
	}
	if n > 1 {
		return Options{}, errors.New("password sources -p/-e/-f/-d are mutually exclusive")
	}

	return opt, nil
}
