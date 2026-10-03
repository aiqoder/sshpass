package prompt

import "testing"

func TestPasswordMatch(t *testing.T) {
	m := New("assword", "are you sure you want to continue connecting", true)
	ok, rest := m.PasswordMatched([]byte("Login Password: "))
	if !ok {
		t.Fatal("expected password match")
	}
	if string(rest) != ": " {
		t.Fatalf("rest=%q", rest)
	}
}

func TestHostThenPassword(t *testing.T) {
	m := New("assword", "are you sure you want to continue connecting (yes/no", true)
	buf := []byte("The authenticity of host 'x' can't be established.\nAre you sure you want to continue connecting (yes/no)? ")
	ok, rest := m.HostMatched(buf)
	if !ok {
		t.Fatal("expected host match")
	}
	ok, _ = m.PasswordMatched(rest)
	if ok {
		t.Fatal("password should not match host prompt remainder")
	}

	buf2 := append(rest, []byte("\nPassword: ")...)
	ok, _ = m.PasswordMatched(buf2)
	if !ok {
		t.Fatal("expected password match after host")
	}
}

func TestHostDisabled(t *testing.T) {
	m := New("assword", "are you sure you want to continue connecting (yes/no", false)
	ok, _ := m.HostMatched([]byte("Are you sure you want to continue connecting (yes/no)? "))
	if ok {
		t.Fatal("host matching should be disabled")
	}
}

func TestTrimTail(t *testing.T) {
	got := TrimTail([]byte("abcdef"), 3)
	if string(got) != "def" {
		t.Fatalf("got %q", got)
	}
}
