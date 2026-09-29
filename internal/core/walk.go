package core

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// walkers is how many directories are read at once. Reading in parallel
// makes large trees several times faster on NTFS and SSDs.
var walkers = min(8, max(2, runtime.NumCPU()))

// walk collects qualifying files under root. It never follows links, never
// leaves the root's filesystem and skips anything it cannot read.
func walk(ctx context.Context, t *Target, root string, now time.Time, f *found) error {
	w := &walker{ctx: ctx, t: t, root: root, now: now, queue: []string{""}, pending: 1}
	w.cond = sync.NewCond(&w.mu)
	if info, err := os.Stat(root); err == nil {
		w.rootDev, w.haveDev = deviceID(info)
	}
	var wg sync.WaitGroup
	for range walkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.work()
		}()
	}
	wg.Wait()

	sort.Slice(w.files, func(i, j int) bool { return w.files[i].Rel < w.files[j].Rel })
	f.files = append(f.files, w.files...)
	f.oldDirs = append(f.oldDirs, w.oldDirs...)
	f.partial = f.partial || w.partial
	return ctx.Err()
}

type walker struct {
	ctx     context.Context
	t       *Target
	root    string
	now     time.Time
	rootDev uint64
	haveDev bool

	mu   sync.Mutex
	cond *sync.Cond
	// queue holds directories to read, relative to root.
	queue []string
	// pending counts directories queued or being read.
	pending int

	files   []Candidate
	oldDirs []Candidate
	partial bool
}

func (w *walker) work() {
	for {
		w.mu.Lock()
		for len(w.queue) == 0 && w.pending > 0 {
			w.cond.Wait()
		}
		if w.pending == 0 {
			w.mu.Unlock()
			return
		}
		dir := w.queue[len(w.queue)-1]
		w.queue = w.queue[:len(w.queue)-1]
		w.mu.Unlock()

		var r dirResult
		if w.ctx.Err() == nil {
			r = w.read(dir)
		}

		w.mu.Lock()
		w.queue = append(w.queue, r.subdirs...)
		w.pending += len(r.subdirs) - 1
		w.files = append(w.files, r.files...)
		w.oldDirs = append(w.oldDirs, r.oldDirs...)
		w.partial = w.partial || r.partial
		w.mu.Unlock()
		w.cond.Broadcast()
	}
}

type dirResult struct {
	subdirs []string
	files   []Candidate
	oldDirs []Candidate
	partial bool
}

// read lists one directory. Entries are examined without following links:
// symlinks, junctions and other reparse points are neither entered nor
// collected.
func (w *walker) read(dir string) (r dirResult) {
	d, err := os.Open(filepath.Join(w.root, dir))
	if err != nil {
		r.partial = !os.IsNotExist(err)
		return r
	}
	entries, err := d.ReadDir(-1)
	d.Close()
	if err != nil {
		r.partial = true
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || isLink(info) {
			continue // vanished, or a link we must not follow
		}
		rel := filepath.Join(dir, e.Name())
		if info.IsDir() {
			if w.haveDev {
				if dev, ok := deviceID(info); ok && dev != w.rootDev {
					continue // another filesystem is mounted here
				}
			}
			if w.now.Sub(info.ModTime()) >= w.t.MinAge {
				r.oldDirs = append(r.oldDirs, Candidate{Root: w.root, Rel: rel, ModTime: info.ModTime()})
			}
			r.subdirs = append(r.subdirs, rel)
			continue
		}
		if !info.Mode().IsRegular() || w.now.Sub(info.ModTime()) < w.t.MinAge {
			continue
		}
		if w.t.Match != nil && !w.t.Match(filepath.ToSlash(rel)) {
			continue
		}
		r.files = append(r.files, Candidate{Root: w.root, Rel: rel, Size: info.Size(), ModTime: info.ModTime()})
	}
	return r
}
