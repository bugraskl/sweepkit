package targets

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/bugraskl/sweepkit/internal/core"
)

// keepNewest selects the package files to delete so that only the n most
// recently downloaded versions of each package remain. Signature files go
// with their package. Files that do not look like packages are kept.
func keepNewest(n int) func([]core.Candidate) []core.Candidate {
	return func(found []core.Candidate) []core.Candidate {
		groups := map[string][]core.Candidate{}
		sigs := map[string]core.Candidate{}
		for _, c := range found {
			if strings.HasSuffix(c.Rel, ".sig") {
				sigs[strings.TrimSuffix(c.Rel, ".sig")] = c
				continue
			}
			if key, ok := packageKey(filepath.Base(c.Rel)); ok {
				groups[key] = append(groups[key], c)
			}
		}
		var out []core.Candidate
		for _, versions := range groups {
			sort.Slice(versions, func(i, j int) bool { return versions[i].ModTime.After(versions[j].ModTime) })
			for _, old := range versions[min(n, len(versions)):] {
				out = append(out, old)
				if sig, ok := sigs[old.Rel]; ok {
					out = append(out, sig)
				}
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Rel < out[j].Rel })
		return out
	}
}

// packageKey returns the name and architecture of a package file named
// name-pkgver-pkgrel-arch.pkg.tar.zst. Names may contain dashes; the last
// three fields never do.
func packageKey(file string) (string, bool) {
	stem, _, ok := strings.Cut(file, ".pkg.tar")
	if !ok {
		return "", false
	}
	fields := strings.Split(stem, "-")
	if len(fields) < 4 {
		return "", false
	}
	name := strings.Join(fields[:len(fields)-3], "-")
	return name + " " + fields[len(fields)-1], true
}
