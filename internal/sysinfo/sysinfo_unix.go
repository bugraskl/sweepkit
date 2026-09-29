//go:build !windows

package sysinfo

import (
	"os"
	"os/exec"
	"strings"
)

// Elevated reports whether the process runs as root.
func Elevated() bool { return os.Geteuid() == 0 }

// processNames reads process names from /proc. On systems without /proc it
// returns an error, and Running then reports nothing as running.
func processNames() (map[string]bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() || strings.TrimLeft(e.Name(), "0123456789") != "" {
			continue
		}
		comm, err := os.ReadFile("/proc/" + e.Name() + "/comm")
		if err != nil {
			continue
		}
		names[strings.ToLower(strings.TrimSpace(string(comm)))] = true
	}
	return names, nil
}

func hideWindow(*exec.Cmd) {}
