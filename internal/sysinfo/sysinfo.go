// Package sysinfo answers questions about the running system: privileges,
// running programs and installed commands.
package sysinfo

import (
	"context"
	"os/exec"
	"strings"
)

// Running returns the names from the list that belong to a running
// process. Names are compared without regard to case.
func Running(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	procs, err := processNames()
	if err != nil {
		return nil
	}
	var out []string
	for _, n := range names {
		if procs[strings.ToLower(n)] {
			out = append(out, n)
		}
	}
	return out
}

// Run runs a command and returns its combined output.
func Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	return cmd.CombinedOutput()
}
