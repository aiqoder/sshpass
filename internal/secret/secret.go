package secret

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const EnvName = "SSHPASS"

type Source struct {
	Password string
	UseEnv   bool
	File     string
	FD       int
}

func Read(src Source) (string, error) {
	var (
		pass string
		err  error
	)
	switch {
	case src.Password != "":
		pass = src.Password
	case src.UseEnv:
		pass, err = fromEnv()
	case src.File != "":
		pass, err = fromFile(src.File)
	case src.FD >= 0:
		pass, err = fromFD(src.FD)
	default:
		return "", errors.New("no password source")
	}
	if err != nil {
		return "", err
	}
	pass = strings.TrimRight(pass, "\r\n")
	if pass == "" {
		return "", errors.New("password is empty")
	}
	return pass, nil
}

func fromEnv() (string, error) {
	v, ok := os.LookupEnv(EnvName)
	if !ok {
		return "", fmt.Errorf("environment variable %s is not set", EnvName)
	}
	return v, nil
}

func fromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open password file: %w", err)
	}
	defer f.Close()
	return firstLine(f)
}

func fromFD(fd int) (string, error) {
	f := os.NewFile(uintptr(fd), "password-fd")
	if f == nil {
		return "", fmt.Errorf("invalid file descriptor %d", fd)
	}
	if fd > 2 {
		defer f.Close()
	}
	return firstLine(f)
}

func firstLine(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read password: %w", err)
	}
	if line == "" && errors.Is(err, io.EOF) {
		return "", errors.New("password is empty")
	}
	return line, nil
}
