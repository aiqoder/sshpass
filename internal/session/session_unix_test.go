//go:build unix

package session

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunHostThenPassword(t *testing.T) {
	script := `
printf '%s' "Are you sure you want to continue connecting (yes/no)? "
read ans
printf 'got-host:%s\n' "$ans"
printf '%s' "Password: "
read pass
printf 'got-pass:%s\n' "$pass"
exit 7
`
	var stdout bytes.Buffer
	code, err := Run(context.Background(), Options{
		Command:        []string{"sh", "-c", script},
		Password:       "s3cret",
		PasswordPrompt: "assword",
		HostPrompt:     "are you sure you want to continue connecting (yes/no",
		HostConfirm:    true,
		Timeout:        5 * time.Second,
		Stdout:         &stdout,
		Stderr:         &bytes.Buffer{},
		Stdin:          bytes.NewReader(nil),
	})
	if err != nil {
		t.Fatalf("run: %v\noutput:\n%s", err, stdout.String())
	}
	if code != 7 {
		t.Fatalf("exit code=%d output=%q", code, stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "got-host:yes") {
		t.Fatalf("missing host answer in %q", out)
	}
	if !strings.Contains(out, "got-pass:s3cret") {
		t.Fatalf("missing password in %q", out)
	}
}
