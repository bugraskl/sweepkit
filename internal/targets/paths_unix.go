//go:build !windows

package targets

import (
	"bufio"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func detect() Paths {
	// Under sudo, clean the invoking user's caches rather than root's, and
	// ignore environment variables, which may belong to either user.
	home, sudo := invokingHome()
	getenv := os.Getenv
	if sudo {
		getenv = func(string) string { return "" }
	}
	xdg := func(key, fallback string) string {
		return firstAbs(getenv(key), join(home, fallback))
	}
	p := Paths{
		Home:       home,
		CacheHome:  xdg("XDG_CACHE_HOME", ".cache"),
		ConfigHome: xdg("XDG_CONFIG_HOME", ".config"),
		DataHome:   xdg("XDG_DATA_HOME", ".local/share"),
		Getenv:     getenv,
	}
	goCache := ""
	if !sudo {
		goCache = firstAbs(os.Getenv("GOCACHE"), goEnvCache())
	}
	p.GoCache = firstAbs(goCache, join(p.CacheHome, "go-build"))

	for _, name := range []string{"Documents", "Desktop", "Downloads", "Pictures", "Music", "Videos", "Public", "Templates"} {
		p.Personal = append(p.Personal, join(home, name))
	}
	p.Personal = append(p.Personal, userDirs(home, join(p.ConfigHome, "user-dirs.dirs"))...)
	return p
}

// invokingHome returns the home folder of the user who ran the program,
// looking through sudo.
func invokingHome() (home string, sudo bool) {
	if os.Geteuid() == 0 {
		if name := os.Getenv("SUDO_USER"); name != "" && name != "root" {
			if u, err := user.Lookup(name); err == nil && filepath.IsAbs(u.HomeDir) {
				return u.HomeDir, true
			}
		}
	}
	home, _ = os.UserHomeDir()
	return home, false
}

// userDirs reads the localized document folders from user-dirs.dirs, e.g.
// XDG_DOCUMENTS_DIR="$HOME/Belgeler".
func userDirs(home, file string) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()
	var dirs []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(s.Text()), "=")
		if !ok || !strings.HasPrefix(key, "XDG_") || !strings.HasSuffix(key, "_DIR") {
			continue
		}
		value = strings.Trim(value, `"`)
		value = strings.Replace(value, "$HOME", home, 1)
		if filepath.IsAbs(value) && filepath.Clean(value) != filepath.Clean(home) {
			dirs = append(dirs, filepath.Clean(value))
		}
	}
	return dirs
}
