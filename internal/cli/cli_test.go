package cli

import (
	"io"
	"testing"
	"time"
)

func TestParsePasswordSourcesExclusive(t *testing.T) {
	_, err := Parse([]string{"-p", "x", "-e", "--", "ssh", "h"}, io.Discard)
	if err == nil {
		t.Fatal("expected exclusive source error")
	}
}

func TestParseMissingCommand(t *testing.T) {
	_, err := Parse([]string{"-e"}, io.Discard)
	if err == nil {
		t.Fatal("expected missing command")
	}
}

func TestParseOK(t *testing.T) {
	opt, err := Parse([]string{"-e", "-t", "10", "--no-host-confirm", "--", "ssh", "user@host"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !opt.UseEnv || opt.HostConfirm || opt.Timeout != 10*time.Second {
		t.Fatalf("unexpected opt: %+v", opt)
	}
	if len(opt.Command) != 2 || opt.Command[0] != "ssh" {
		t.Fatalf("command: %v", opt.Command)
	}
}

func TestParseDefaultHostConfirm(t *testing.T) {
	opt, err := Parse([]string{"-p", "secret", "sudo", "-S", "id"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !opt.HostConfirm {
		t.Fatal("host confirm should be on by default")
	}
}

func TestParseMissingPasswordSource(t *testing.T) {
	_, err := Parse([]string{"ssh", "host"}, io.Discard)
	if err == nil {
		t.Fatal("expected missing password source")
	}
}
