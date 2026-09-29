package targets

import (
	"path/filepath"

	"github.com/bugraskl/sweepkit/internal/core"
)

// GuardFor lists the folders that must never be cleaned on this system.
func GuardFor(p Paths) *core.Guard {
	g := &core.Guard{Sealed: nonEmpty(p.Personal...)}
	switch p.OS {
	case "windows":
		g.Sealed = append(g.Sealed, nonEmpty(
			join(p.SystemRoot, "System32"),
			join(p.SystemRoot, "SysWOW64"),
			join(p.SystemRoot, "WinSxS"),
			join(p.SystemRoot, "Installer"),
			join(p.SystemRoot, "servicing"),
			join(p.SystemRoot, "Boot"),
			join(p.SystemRoot, "Fonts"),
			p.ProgramFiles,
			p.ProgramFilesX86,
		)...)
		g.Protected = nonEmpty(
			p.SystemRoot,
			p.ProgramData,
			join(p.ProgramData, "Microsoft"),
			join(p.ProgramData, "Microsoft", "Windows"),
			p.Home,
			parent(p.Home),
			join(p.Home, "AppData"),
			p.LocalAppData,
			p.RoamingAppData,
			p.LocalLow,
			join(p.LocalAppData, "Microsoft"),
			join(p.LocalAppData, "Microsoft", "Windows"),
		)
	default:
		g.Sealed = append(g.Sealed,
			"/bin", "/boot", "/dev", "/etc", "/lib", "/lib32", "/lib64", "/libx32",
			"/opt", "/proc", "/run", "/sbin", "/snap", "/srv", "/sys", "/usr",
			"/var/lib/apt", "/var/lib/dpkg", "/var/lib/pacman", "/var/lib/rpm", "/var/lib/dnf",
		)
		g.Protected = append(nonEmpty(
			p.Home,
			p.CacheHome,
			p.ConfigHome,
			p.DataHome,
			join(p.Home, ".local"),
			join(p.Home, "snap"),
			join(p.Home, ".var", "app"),
		), "/home", "/root", "/var", "/var/lib", "/var/cache", "/var/log", "/tmp", "/var/tmp", "/mnt", "/media")
	}
	return g
}

func parent(p string) string {
	if p == "" {
		return ""
	}
	return filepath.Dir(p)
}
