package sysinfo

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

// Only the standard syscall package is used here, so the command-line tool
// does not link any networking code (golang.org/x/sys/windows would).

const (
	tokenElevation = 20         // TOKEN_INFORMATION_CLASS TokenElevation
	createNoWindow = 0x08000000 // CREATE_NO_WINDOW
)

// Elevated reports whether the process runs with administrator rights.
func Elevated() bool {
	proc, err := syscall.GetCurrentProcess()
	if err != nil {
		return false
	}
	var token syscall.Token
	if syscall.OpenProcessToken(proc, syscall.TOKEN_QUERY, &token) != nil {
		return false
	}
	defer token.Close()
	var elevated, n uint32
	if syscall.GetTokenInformation(token, tokenElevation, (*byte)(unsafe.Pointer(&elevated)), uint32(unsafe.Sizeof(elevated)), &n) != nil {
		return false
	}
	return elevated != 0
}

func processNames() (map[string]bool, error) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(snap)

	names := map[string]bool{}
	var e syscall.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = syscall.Process32First(snap, &e); err == nil; err = syscall.Process32Next(snap, &e) {
		names[strings.ToLower(syscall.UTF16ToString(e.ExeFile[:]))] = true
	}
	return names, nil
}

// hideWindow keeps console programs from flashing a window when started
// from the desktop app.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
