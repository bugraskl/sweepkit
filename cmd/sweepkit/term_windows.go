package main

import (
	"os"
	"syscall"
)

// isTerminal reports whether f is a console. It uses the standard syscall
// package so the CLI links no networking code.
func isTerminal(f *os.File) bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode) == nil
}
