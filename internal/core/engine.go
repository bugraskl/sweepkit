package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// maxErrors caps how many error messages a CleanResult carries.
const maxErrors = 5

// found is what a walk over a target's locations turned up.
type found struct {
	files []Candidate
	// oldDirs are directories untouched for MinAge. They are removed after
	// cleaning if nothing is left in them.
	oldDirs []Candidate
	// exists is true if at least one location exists.
	exists   bool
	partial  bool
	warnings []string
}

func (f *found) totals() (files int, bytes int64) {
	for _, c := range f.files {
		bytes += c.Size
	}
	return len(f.files), bytes
}

// Present reports whether any of the target's locations exist and, for
// command targets, whether the command is installed. It is cheap enough to
// call for every target when listing them.
func Present(env *Env, t *Target) bool {
	if len(t.Command) > 0 {
		if _, err := env.LookPath(t.Command[0]); err != nil {
			return false
		}
	}
	for _, p := range append(append([]string{}, t.Roots...), t.Files...) {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// Scan reports what cleaning t would remove. It never modifies anything.
func Scan(ctx context.Context, env *Env, t *Target) ScanResult {
	r := ScanResult{ID: t.ID}
	if len(t.Command) > 0 {
		if _, err := env.LookPath(t.Command[0]); err != nil {
			r.Status = NotFound
			return r
		}
	}
	f, err := find(ctx, env, t)
	if err != nil {
		r.Status, r.Error = Failed, err.Error()
		return r
	}
	r.Warnings, r.Partial = f.warnings, f.partial
	if !f.exists {
		r.Status = NotFound
		return r
	}
	r.Files, r.Bytes = f.totals()
	r.Status, r.Blockers = readiness(env, t)
	return r
}

// Clean removes what Scan would report, checking every file again right
// before it is deleted.
func Clean(ctx context.Context, env *Env, t *Target) CleanResult {
	r := CleanResult{ID: t.ID}
	if len(t.Command) > 0 {
		if _, err := env.LookPath(t.Command[0]); err != nil {
			r.Status = NotFound
			return r
		}
	}
	f, err := find(ctx, env, t)
	if err != nil {
		r.Status, r.Error = Failed, err.Error()
		return r
	}
	if !f.exists {
		r.Status = NotFound
		return r
	}
	if r.Status, r.Blockers = readiness(env, t); r.Status != Ready {
		return r
	}
	if len(t.Command) > 0 {
		return runCommand(ctx, env, t, f, r)
	}
	deleteAll(ctx, env, t, f, &r)
	r.Status = Done
	if ctx.Err() != nil {
		r.Status, r.Error = Failed, ctx.Err().Error()
	}
	return r
}

func readiness(env *Env, t *Target) (Status, []string) {
	if t.Admin && !env.Elevated {
		return NeedsAdmin, nil
	}
	if len(t.BlockedBy) > 0 {
		if running := env.Running(t.BlockedBy); len(running) > 0 {
			return Blocked, running
		}
	}
	return Ready, nil
}

// find walks the target's locations and collects every file that qualifies.
func find(ctx context.Context, env *Env, t *Target) (*found, error) {
	f := &found{}
	now := env.Now()
	for _, root := range t.Roots {
		if _, err := os.Stat(root); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				f.exists, f.partial = true, true
			}
			continue
		}
		resolved, err := env.Guard.CheckRoot(root)
		if errors.Is(err, fs.ErrPermission) {
			// Readable only with more rights; the target reports it needs them.
			f.exists, f.partial = true, true
			continue
		}
		if err != nil {
			f.warnings = append(f.warnings, err.Error())
			continue
		}
		f.exists = true
		if err := walk(ctx, t, resolved, now, f); err != nil {
			return nil, err
		}
	}
	for _, file := range t.Files {
		info, err := os.Lstat(file)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				f.exists, f.partial = true, true
			}
			continue
		}
		resolved, err := env.Guard.CheckFile(file)
		if err != nil {
			f.warnings = append(f.warnings, err.Error())
			continue
		}
		f.exists = true
		if !info.Mode().IsRegular() || isLink(info) || now.Sub(info.ModTime()) < t.MinAge {
			continue
		}
		f.files = append(f.files, Candidate{
			Root: filepath.Dir(resolved), Rel: filepath.Base(resolved),
			Size: info.Size(), ModTime: info.ModTime(),
		})
	}
	if t.Select != nil {
		f.files = t.Select(f.files)
	}
	return f, nil
}

// deleteAll removes every candidate through an os.Root opened on its
// directory, so a folder swapped for a link after the scan cannot redirect
// a deletion outside that directory.
func deleteAll(ctx context.Context, env *Env, t *Target, f *found, r *CleanResult) {
	var roots []string
	seen := map[string]bool{}
	addRoot := func(root string) {
		if !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}
	byRoot := map[string][]Candidate{}
	for _, c := range f.files {
		addRoot(c.Root)
		byRoot[c.Root] = append(byRoot[c.Root], c)
	}
	dirsByRoot := map[string][]Candidate{}
	if t.Match == nil && t.Select == nil {
		// Only tidy up folders when the whole tree is ours to clean.
		for _, d := range f.oldDirs {
			addRoot(d.Root)
			dirsByRoot[d.Root] = append(dirsByRoot[d.Root], d)
		}
	}
	now := env.Now()
	for _, root := range roots {
		dir, err := os.OpenRoot(root)
		if err != nil {
			if len(byRoot[root]) > 0 {
				r.Failed += len(byRoot[root])
				r.addError(err)
			}
			continue
		}
		for _, c := range byRoot[root] {
			if ctx.Err() != nil {
				break
			}
			deleteOne(dir, c, t, now, r)
		}
		removeEmptyDirs(dir, dirsByRoot[root])
		dir.Close()
	}
}

func deleteOne(dir *os.Root, c Candidate, t *Target, now time.Time, r *CleanResult) {
	info, err := dir.Lstat(c.Rel)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return // already gone
	case err != nil:
		r.Failed++
		r.addError(err)
		return
	case !info.Mode().IsRegular() || isLink(info) || now.Sub(info.ModTime()) < t.MinAge:
		return // changed since the scan; leave it alone
	}
	err = dir.Remove(c.Rel)
	if err != nil && runtime.GOOS == "windows" && info.Mode().Perm()&0o200 == 0 {
		// Clear the read-only attribute and try once more.
		if dir.Chmod(c.Rel, 0o666) == nil {
			err = dir.Remove(c.Rel)
		}
	}
	switch {
	case err == nil:
		r.Deleted++
		r.Freed += info.Size()
	case errors.Is(err, fs.ErrNotExist):
	case isInUse(err):
		r.InUse++
	default:
		r.Failed++
		r.addError(err)
	}
}

// removeEmptyDirs removes directories, deepest first. Removing a directory
// that still has files in it fails, which is what we want.
func removeEmptyDirs(dir *os.Root, dirs []Candidate) {
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i].Rel, string(filepath.Separator)) > strings.Count(dirs[j].Rel, string(filepath.Separator))
	})
	for _, d := range dirs {
		if info, err := dir.Lstat(d.Rel); err != nil || !info.IsDir() || isLink(info) {
			continue
		}
		_ = dir.Remove(d.Rel)
	}
}

// runCommand cleans with the target's tool and measures the difference.
func runCommand(ctx context.Context, env *Env, t *Target, before *found, r CleanResult) CleanResult {
	out, err := env.Run(ctx, t.Command[0], t.Command[1:]...)
	if err != nil {
		r.Status = Failed
		r.Error = fmt.Sprintf("%s: %v", strings.Join(t.Command, " "), err)
		if msg := strings.TrimSpace(string(out)); msg != "" {
			r.Error += ": " + lastLine(msg)
		}
		return r
	}
	after, err := find(ctx, env, t)
	if err != nil {
		r.Status, r.Error = Done, "cleaned, but could not measure the result: "+err.Error()
		return r
	}
	n0, b0 := before.totals()
	n1, b1 := after.totals()
	r.Status = Done
	r.Deleted, r.Freed = max(n0-n1, 0), max(b0-b1, 0)
	return r
}

func (r *CleanResult) addError(err error) {
	if len(r.Errors) < maxErrors {
		r.Errors = append(r.Errors, err.Error())
	}
}

func lastLine(s string) string {
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return s
}
