//go:build windows

package session

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

type winProc struct {
	in   *os.File
	out  *os.File
	pc   windows.Handle
	proc windows.Handle
	env  []uint16
}

func startPTY(name string, args []string) (ptyProcess, error) {
	exe, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("look up %q: %w", name, err)
	}

	cmdLine := windows.ComposeCommandLine(append([]string{exe}, args...))

	var ptyIn, ptyOut, ourIn, ourOut windows.Handle
	if err := windows.CreatePipe(&ptyIn, &ourIn, nil, 0); err != nil {
		return nil, fmt.Errorf("create input pipe: %w", err)
	}
	if err := windows.CreatePipe(&ourOut, &ptyOut, nil, 0); err != nil {
		closeHandles(ptyIn, ourIn)
		return nil, fmt.Errorf("create output pipe: %w", err)
	}

	size := windows.Coord{X: 80, Y: 24}
	var hPC windows.Handle
	if err := windows.CreatePseudoConsole(size, ptyIn, ptyOut, 0, &hPC); err != nil {
		closeHandles(ptyIn, ptyOut, ourIn, ourOut)
		return nil, fmt.Errorf("CreatePseudoConsole: %w", err)
	}

	_ = windows.CloseHandle(ptyIn)
	_ = windows.CloseHandle(ptyOut)

	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		windows.ClosePseudoConsole(hPC)
		closeHandles(ourIn, ourOut)
		return nil, fmt.Errorf("attribute list: %w", err)
	}
	if err := attrList.Update(windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, unsafe.Pointer(hPC), unsafe.Sizeof(hPC)); err != nil {
		attrList.Delete()
		windows.ClosePseudoConsole(hPC)
		closeHandles(ourIn, ourOut)
		return nil, fmt.Errorf("update attribute: %w", err)
	}

	var si windows.StartupInfoEx
	si.Cb = uint32(unsafe.Sizeof(si))
	si.ProcThreadAttributeList = attrList.List()

	envKeep := utf16EnvBlock(os.Environ())
	var pi windows.ProcessInformation
	flags := uint32(windows.CREATE_UNICODE_ENVIRONMENT | windows.EXTENDED_STARTUPINFO_PRESENT)
	cmdPtr, err := windows.UTF16PtrFromString(cmdLine)
	if err != nil {
		attrList.Delete()
		windows.ClosePseudoConsole(hPC)
		closeHandles(ourIn, ourOut)
		return nil, fmt.Errorf("command line: %w", err)
	}

	var envPtr *uint16
	if len(envKeep) > 0 {
		envPtr = &envKeep[0]
	}

	err = windows.CreateProcess(
		nil,
		cmdPtr,
		nil,
		nil,
		false,
		flags,
		envPtr,
		nil,
		&si.StartupInfo,
		&pi,
	)
	attrList.Delete()
	if err != nil {
		windows.ClosePseudoConsole(hPC)
		closeHandles(ourIn, ourOut)
		return nil, fmt.Errorf("CreateProcess: %w", err)
	}
	_ = windows.CloseHandle(pi.Thread)

	return &winProc{
		in:   os.NewFile(uintptr(ourIn), "conpty-in"),
		out:  os.NewFile(uintptr(ourOut), "conpty-out"),
		pc:   hPC,
		proc: pi.Process,
		env:  envKeep,
	}, nil
}

func utf16EnvBlock(env []string) []uint16 {
	n := 1
	for _, e := range env {
		n += len(e) + 1
	}
	buf := make([]uint16, 0, n+8)
	for _, e := range env {
		u, err := windows.UTF16FromString(e)
		if err != nil {
			continue
		}
		buf = append(buf, u[:len(u)-1]...)
		buf = append(buf, 0)
	}
	buf = append(buf, 0)
	return buf
}

func closeHandles(hs ...windows.Handle) {
	for _, h := range hs {
		if h != 0 {
			_ = windows.CloseHandle(h)
		}
	}
}

func (p *winProc) Read(b []byte) (int, error)  { return p.out.Read(b) }
func (p *winProc) Write(b []byte) (int, error) { return p.in.Write(b) }

func (p *winProc) Close() error {
	if p.in != nil {
		_ = p.in.Close()
		p.in = nil
	}
	if p.out != nil {
		_ = p.out.Close()
		p.out = nil
	}
	if p.pc != 0 {
		windows.ClosePseudoConsole(p.pc)
		p.pc = 0
	}
	if p.proc != 0 {
		_ = windows.CloseHandle(p.proc)
		p.proc = 0
	}
	return nil
}

func (p *winProc) Wait() (int, error) {
	if p.proc == 0 {
		return 1, fmt.Errorf("process handle closed")
	}
	s, err := windows.WaitForSingleObject(p.proc, windows.INFINITE)
	if err != nil {
		return 1, err
	}
	if s != windows.WAIT_OBJECT_0 {
		return 1, fmt.Errorf("WaitForSingleObject: %d", s)
	}
	var code uint32
	if err := windows.GetExitCodeProcess(p.proc, &code); err != nil {
		return 1, err
	}
	return int(code), nil
}

func (p *winProc) Kill() error {
	if p.proc == 0 {
		return nil
	}
	return windows.TerminateProcess(p.proc, 1)
}

func (p *winProc) Resize(rows, cols uint16) error {
	if p.pc == 0 {
		return fmt.Errorf("console closed")
	}
	return windows.ResizePseudoConsole(p.pc, windows.Coord{X: int16(cols), Y: int16(rows)})
}
