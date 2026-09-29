package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ErrRefused is returned for locations the guard will not touch.
var ErrRefused = errors.New("refused: location holds personal or system files")

// Guard is the last line of defence against a misconfigured target or a
// hostile environment variable pointing a cleaner at the wrong folder.
type Guard struct {
	// Sealed paths and everything beneath them are never touched.
	Sealed []string
	// Protected paths may contain cleanable directories, but neither they
	// nor any of their parents may be cleaned as a whole.
	Protected []string

	once              sync.Once
	sealed, protected []string
}

// lists returns Sealed and Protected together with the real location of
// each entry, so links and short names cannot slip past a comparison.
func (g *Guard) lists() (sealed, protected []string) {
	g.once.Do(func() {
		g.sealed = withResolved(g.Sealed)
		g.protected = withResolved(g.Protected)
	})
	return g.sealed, g.protected
}

func withResolved(paths []string) []string {
	out := append([]string{}, paths...)
	for _, p := range paths {
		if r, err := resolve(p); err == nil && normalize(r) != normalize(p) {
			out = append(out, r)
		}
	}
	return out
}

// CheckRoot reports whether dir may be used as a root. Links in dir are
// resolved first, so a link cannot smuggle a protected folder past the
// check. It returns the resolved path, which is what must be cleaned.
func (g *Guard) CheckRoot(dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("%s: not an absolute path", dir)
	}
	clean := filepath.Clean(dir)
	resolved, err := resolve(clean)
	if err != nil {
		return "", err
	}
	sealed, protected := g.lists()
	for _, p := range []string{clean, resolved} {
		if isVolumeRoot(p) {
			return "", fmt.Errorf("%s: %w", dir, ErrRefused)
		}
		for _, s := range sealed {
			if within(p, s) || within(s, p) {
				return "", fmt.Errorf("%s: %w", dir, ErrRefused)
			}
		}
		for _, q := range protected {
			if within(q, p) {
				return "", fmt.Errorf("%s: %w", dir, ErrRefused)
			}
		}
	}
	return resolved, nil
}

// CheckFile reports whether a single named file may be removed. It returns
// the file's path with its parent directory resolved.
func (g *Guard) CheckFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s: not an absolute path", path)
	}
	clean := filepath.Clean(path)
	parent, err := resolve(filepath.Dir(clean))
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(parent, filepath.Base(clean))
	sealed, _ := g.lists()
	for _, p := range []string{clean, resolved} {
		for _, s := range sealed {
			if within(p, s) {
				return "", fmt.Errorf("%s: %w", path, ErrRefused)
			}
		}
	}
	return resolved, nil
}

// within reports whether path is dir or lies beneath it.
func within(path, dir string) bool {
	p, d := normalize(path), normalize(dir)
	if p == d {
		return true
	}
	if !strings.HasSuffix(d, string(filepath.Separator)) {
		d += string(filepath.Separator)
	}
	return strings.HasPrefix(p, d)
}

func normalize(p string) string {
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

// isVolumeRoot reports whether p is a filesystem root such as "/", "D:\"
// or a network share.
func isVolumeRoot(p string) bool {
	p = filepath.Clean(p)
	vol := filepath.VolumeName(p)
	return p == vol || p == vol+string(filepath.Separator)
}
