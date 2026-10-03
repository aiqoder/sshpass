package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aiqoder/sshpass/internal/cli"
	"github.com/aiqoder/sshpass/internal/secret"
	"github.com/aiqoder/sshpass/internal/session"
)

func main() {
	opt, err := cli.Parse(os.Args[1:], os.Stderr)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "sshpass:", err)
		os.Exit(1)
	}

	pass, err := secret.Read(secret.Source{
		Password: opt.Password,
		UseEnv:   opt.UseEnv,
		File:     opt.File,
		FD:       opt.FD,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "sshpass:", err)
		os.Exit(1)
	}

	code, err := session.Run(context.Background(), session.Options{
		Command:        opt.Command,
		Password:       pass,
		PasswordPrompt: opt.PasswordPrompt,
		HostPrompt:     opt.HostPrompt,
		HostConfirm:    opt.HostConfirm,
		Timeout:        opt.Timeout,
		Verbose:        opt.Verbose,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "sshpass:", err)
	}
	os.Exit(code)
}
