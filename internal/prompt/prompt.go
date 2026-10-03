package prompt

import "bytes"

type Matcher struct {
	PasswordNeedle []byte
	HostNeedle     []byte
	HostConfirm    bool
}

func New(passwordPrompt, hostPrompt string, hostConfirm bool) Matcher {
	return Matcher{
		PasswordNeedle: bytes.ToLower([]byte(passwordPrompt)),
		HostNeedle:     bytes.ToLower([]byte(hostPrompt)),
		HostConfirm:    hostConfirm,
	}
}

func (m Matcher) HostMatched(buf []byte) (matched bool, rest []byte) {
	if !m.HostConfirm || len(m.HostNeedle) == 0 {
		return false, buf
	}
	return cutMatch(buf, m.HostNeedle)
}

func (m Matcher) PasswordMatched(buf []byte) (matched bool, rest []byte) {
	if len(m.PasswordNeedle) == 0 {
		return false, buf
	}
	return cutMatch(buf, m.PasswordNeedle)
}

func cutMatch(buf, needle []byte) (bool, []byte) {
	lower := bytes.ToLower(buf)
	i := bytes.Index(lower, needle)
	if i < 0 {
		return false, buf
	}
	end := i + len(needle)
	if end > len(buf) {
		end = len(buf)
	}
	return true, buf[end:]
}

func TrimTail(buf []byte, max int) []byte {
	if max <= 0 || len(buf) <= max {
		return buf
	}
	return buf[len(buf)-max:]
}
