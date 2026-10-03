package secret

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPasswordFlag(t *testing.T) {
	got, err := Read(Source{Password: "abc", FD: -1})
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestReadEmptyPassword(t *testing.T) {
	_, err := Read(Source{Password: "\n", FD: -1})
	if err == nil {
		t.Fatal("expected empty password error")
	}
}

func TestReadEnv(t *testing.T) {
	t.Setenv(EnvName, "from-env")
	got, err := Read(Source{UseEnv: true, FD: -1})
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-env" {
		t.Fatalf("got %q", got)
	}
}

func TestReadEnvMissing(t *testing.T) {
	t.Setenv(EnvName, "")
	_ = os.Unsetenv(EnvName)
	_, err := Read(Source{UseEnv: true, FD: -1})
	if err == nil {
		t.Fatal("expected missing env error")
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.txt")
	if err := os.WriteFile(path, []byte("filepass\nsecond\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(Source{File: path, FD: -1})
	if err != nil {
		t.Fatal(err)
	}
	if got != "filepass" {
		t.Fatalf("got %q", got)
	}
}
