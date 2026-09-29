// Package version reports the sweepkit version.
package version

import "runtime/debug"

// Version is set at build time with
// -ldflags "-X github.com/bugraskl/sweepkit/internal/version.Version=v1.2.3".
var Version = ""

// String returns the version, falling back to the module version recorded
// by `go install`.
func String() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
