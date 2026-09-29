package targets

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bugraskl/sweepkit/internal/core"
)

const (
	hour  = time.Hour
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
)

func txt(en, tr string) core.Text { return core.Text{EN: en, TR: tr} }

// join is filepath.Join that keeps an unknown base unknown instead of
// turning it into a relative path.
func join(base string, elem ...string) string {
	if base == "" {
		return ""
	}
	return filepath.Join(append([]string{base}, elem...)...)
}

func nonEmpty(paths ...string) []string {
	var out []string
	for _, p := range paths {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// dirs expands glob patterns and keeps the directories that exist.
func dirs(patterns ...string) []string {
	var out []string
	seen := map[string]bool{}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		matches, _ := filepath.Glob(pattern)
		for _, m := range matches {
			if info, err := os.Stat(m); err == nil && info.IsDir() && !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	return out
}

func hasExt(ext string) func(string) bool {
	return func(rel string) bool { return strings.EqualFold(filepath.Ext(rel), ext) }
}

// firstDirMatches accepts files whose top-level folder matches re. Caches
// laid out this way stay safe even if their location is misconfigured.
func firstDirMatches(re *regexp.Regexp) func(string) bool {
	return func(rel string) bool {
		first, _, found := strings.Cut(rel, "/")
		return found && re.MatchString(first)
	}
}
