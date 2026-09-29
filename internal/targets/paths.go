// Package targets defines what sweepkit can clean on each operating system.
package targets

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bugraskl/sweepkit/internal/core"
	"github.com/bugraskl/sweepkit/internal/sysinfo"
)

// Paths is the folder layout of the current user and system. Targets are
// built from it, so tests can point them at temporary directories.
type Paths struct {
	OS   string
	Home string

	// Windows.
	LocalAppData    string
	RoamingAppData  string
	LocalLow        string
	SystemRoot      string
	ProgramData     string
	ProgramFiles    string
	ProgramFilesX86 string
	UserTemp        []string

	// Linux XDG base directories.
	CacheHome  string
	ConfigHome string
	DataHome   string

	// Personal holds the user's document folders. Nothing inside them is
	// ever touched.
	Personal []string
	// GoCache is the Go build cache directory.
	GoCache string
	// Getenv reads cache locations that users override with environment
	// variables, such as npm_config_cache.
	Getenv func(string) string
}

// System returns the engine environment and the targets for this computer.
func System() (*core.Env, []*core.Target) {
	p := Detect()
	env := &core.Env{
		Elevated: sysinfo.Elevated(),
		Now:      time.Now,
		Running:  sysinfo.Running,
		LookPath: exec.LookPath,
		Run:      sysinfo.Run,
		Guard:    GuardFor(p),
	}
	return env, For(p)
}

// Detect reads the folder layout of the current user and system.
func Detect() Paths {
	p := detect()
	p.OS = runtime.GOOS
	return p
}

// For returns the targets available for a folder layout. Only Windows and
// Linux are supported; other systems get none.
func For(p Paths) []*core.Target {
	var ts []*core.Target
	switch p.OS {
	case "windows":
		ts = windowsTargets(p)
	case "linux":
		ts = linuxTargets(p)
	default:
		return nil
	}
	ts = append(ts, browserTargets(p)...)
	return append(ts, developerTargets(p)...)
}

// goEnvCache asks the Go toolchain where its build cache is, for users who
// set it with `go env -w` rather than an environment variable.
func goEnvCache() string {
	if _, err := exec.LookPath("go"); err != nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := sysinfo.Run(ctx, "go", "env", "GOCACHE")
	if err != nil {
		return ""
	}
	dir := strings.TrimSpace(string(out))
	if dir == "off" || !filepath.IsAbs(dir) {
		return ""
	}
	return dir
}

func firstAbs(paths ...string) string {
	for _, p := range paths {
		if p != "" && filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
	}
	return ""
}
