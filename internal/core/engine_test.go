package core

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testEnv() *Env {
	return &Env{
		Elevated: true,
		Now:      time.Now,
		Running:  func([]string) []string { return nil },
		LookPath: func(string) (string, error) { return "", errors.New("not installed") },
		Run: func(context.Context, string, ...string) ([]byte, error) {
			return nil, errors.New("unexpected command")
		},
		Guard: &Guard{},
	}
}

// writeFile creates a file of the given size whose modification time lies
// age in the past.
func writeFile(t *testing.T, path string, size int, age time.Duration) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
	setAge(t, path, age)
}

func setAge(t *testing.T, path string, age time.Duration) {
	t.Helper()
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestScanAndCleanRespectMinAge(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "old.tmp"), 100, 48*time.Hour)
	writeFile(t, filepath.Join(root, "nested", "old.tmp"), 50, 48*time.Hour)
	writeFile(t, filepath.Join(root, "new.tmp"), 10, time.Hour)
	target := &Target{ID: "t", Roots: []string{root}, MinAge: 24 * time.Hour}

	s := Scan(context.Background(), testEnv(), target)
	if s.Status != Ready || s.Files != 2 || s.Bytes != 150 {
		t.Fatalf("scan = %+v, want ready with 2 files and 150 bytes", s)
	}
	if !exists(filepath.Join(root, "old.tmp")) {
		t.Fatal("scan deleted a file")
	}

	c := Clean(context.Background(), testEnv(), target)
	if c.Status != Done || c.Deleted != 2 || c.Freed != 150 {
		t.Fatalf("clean = %+v, want 2 files and 150 bytes", c)
	}
	if exists(filepath.Join(root, "old.tmp")) || exists(filepath.Join(root, "nested", "old.tmp")) {
		t.Error("old files survived")
	}
	if !exists(filepath.Join(root, "new.tmp")) {
		t.Error("a recent file was deleted")
	}
	if !exists(root) {
		t.Error("the root itself was deleted")
	}
}

func TestCleanRemovesOnlyAbandonedEmptyDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "abandoned", "deep", "a.log"), 1, 48*time.Hour)
	if err := os.MkdirAll(filepath.Join(root, "fresh"), 0o755); err != nil {
		t.Fatal(err)
	}
	setAge(t, filepath.Join(root, "abandoned", "deep"), 48*time.Hour)
	setAge(t, filepath.Join(root, "abandoned"), 48*time.Hour)

	Clean(context.Background(), testEnv(), &Target{ID: "t", Roots: []string{root}, MinAge: 24 * time.Hour})

	if exists(filepath.Join(root, "abandoned")) {
		t.Error("an emptied, abandoned folder was kept")
	}
	if !exists(filepath.Join(root, "fresh")) {
		t.Error("a recently created folder was removed")
	}
}

func TestMatchLimitsFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.dmp"), 5, 48*time.Hour)
	writeFile(t, filepath.Join(root, "keep.txt"), 5, 48*time.Hour)
	target := &Target{ID: "t", Roots: []string{root}, Match: func(rel string) bool { return strings.HasSuffix(rel, ".dmp") }}

	c := Clean(context.Background(), testEnv(), target)
	if c.Deleted != 1 || exists(filepath.Join(root, "a.dmp")) || !exists(filepath.Join(root, "keep.txt")) {
		t.Fatalf("clean = %+v; Match was not honoured", c)
	}
}

// link makes link point at dir: a symlink on Unix, a junction on Windows,
// where junctions need no special privileges.
func link(t *testing.T, dir, link string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, dir).CombinedOutput(); err != nil {
			t.Fatalf("mklink: %v: %s", err, out)
		}
		return
	}
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
}

func TestLinksAreNeverFollowed(t *testing.T) {
	outside := t.TempDir()
	precious := filepath.Join(outside, "precious.txt")
	writeFile(t, precious, 10, 48*time.Hour)

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "junk.tmp"), 1, 48*time.Hour)
	link(t, outside, filepath.Join(root, "escape"))
	target := &Target{ID: "t", Roots: []string{root}}

	s := Scan(context.Background(), testEnv(), target)
	if s.Files != 1 {
		t.Fatalf("scan found %d files, want only the one inside the root", s.Files)
	}
	Clean(context.Background(), testEnv(), target)
	if !exists(precious) {
		t.Fatal("a file outside the root was deleted through a link")
	}
	if !exists(filepath.Join(root, "escape")) {
		t.Error("the link itself was removed")
	}
}

func TestDeletionCannotEscapeWhenFolderIsSwappedForLink(t *testing.T) {
	outside := t.TempDir()
	precious := filepath.Join(outside, "same-name.tmp")
	writeFile(t, precious, 10, 48*time.Hour)

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "sub", "same-name.tmp"), 1, 48*time.Hour)
	target := &Target{ID: "t", Roots: []string{root}}
	env := testEnv()

	f, err := find(context.Background(), env, target)
	if err != nil || len(f.files) != 1 {
		t.Fatalf("find = %v, %v", f, err)
	}
	// Between scan and delete, an attacker replaces the folder with a link.
	if err := os.RemoveAll(filepath.Join(root, "sub")); err != nil {
		t.Fatal(err)
	}
	link(t, outside, filepath.Join(root, "sub"))

	var r CleanResult
	deleteAll(context.Background(), env, target, f, &r)
	if !exists(precious) {
		t.Fatal("deletion followed a link swapped in after the scan")
	}
}

func TestStatuses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "x"), 1, 48*time.Hour)

	missing := &Target{ID: "t", Roots: []string{filepath.Join(root, "nope")}}
	if s := Scan(context.Background(), testEnv(), missing); s.Status != NotFound {
		t.Errorf("missing root: status %s, want not_found", s.Status)
	}

	env := testEnv()
	env.Elevated = false
	admin := &Target{ID: "t", Roots: []string{root}, Admin: true}
	if s := Scan(context.Background(), env, admin); s.Status != NeedsAdmin || s.Files != 1 {
		t.Errorf("admin target: %+v, want needs_admin with the size still reported", s)
	}
	if c := Clean(context.Background(), env, admin); c.Status != NeedsAdmin || !exists(filepath.Join(root, "x")) {
		t.Errorf("admin target was cleaned without privileges: %+v", c)
	}

	env = testEnv()
	env.Running = func(names []string) []string { return names[:1] }
	browser := &Target{ID: "t", Roots: []string{root}, BlockedBy: []string{"chrome.exe"}}
	if c := Clean(context.Background(), env, browser); c.Status != Blocked || !exists(filepath.Join(root, "x")) {
		t.Errorf("cleaned while the browser was running: %+v", c)
	}
}

func TestGuardRejectsRootsHoldingPersonalOrSystemFiles(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	docs := filepath.Join(home, "Documents")
	cache := filepath.Join(home, "cache")
	for _, d := range []string{docs, filepath.Join(docs, "project"), cache} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	g := &Guard{Sealed: []string{docs}, Protected: []string{home}}

	for _, dir := range []string{home, base, docs, filepath.Join(docs, "project"), filepath.VolumeName(base) + string(filepath.Separator)} {
		if _, err := g.CheckRoot(dir); !errors.Is(err, ErrRefused) {
			t.Errorf("CheckRoot(%s) = %v, want refusal", dir, err)
		}
	}
	if _, err := g.CheckRoot(cache); err != nil {
		t.Errorf("CheckRoot(cache) = %v, want ok", err)
	}
	if _, err := g.CheckRoot("relative/path"); err == nil {
		t.Error("relative root accepted")
	}

	// A link to a sealed folder is judged by where it points.
	link(t, docs, filepath.Join(cache, "sneaky"))
	if _, err := g.CheckRoot(filepath.Join(cache, "sneaky")); !errors.Is(err, ErrRefused) {
		t.Errorf("link to Documents accepted: %v", err)
	}
	if _, err := g.CheckFile(filepath.Join(docs, "notes.txt")); !errors.Is(err, ErrRefused) {
		t.Errorf("file in Documents accepted: %v", err)
	}
}

func TestRefusedRootIsSkippedWithWarning(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "letter.txt"), 1, 48*time.Hour)
	env := testEnv()
	env.Guard = &Guard{Protected: []string{home}}

	c := Clean(context.Background(), env, &Target{ID: "t", Roots: []string{home}})
	if c.Deleted != 0 || !exists(filepath.Join(home, "letter.txt")) {
		t.Fatalf("a protected folder was cleaned: %+v", c)
	}
	if s := Scan(context.Background(), env, &Target{ID: "t", Roots: []string{home}}); len(s.Warnings) == 0 {
		t.Error("no warning about the refused folder")
	}
}

func TestSingleFiles(t *testing.T) {
	dir := t.TempDir()
	dump := filepath.Join(dir, "MEMORY.DMP")
	writeFile(t, dump, 42, 30*24*time.Hour)
	c := Clean(context.Background(), testEnv(), &Target{ID: "t", Files: []string{dump}, MinAge: 7 * 24 * time.Hour})
	if c.Deleted != 1 || c.Freed != 42 || exists(dump) {
		t.Fatalf("clean = %+v", c)
	}
}

func TestCommandTargetMeasuresWhatWasFreed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.deb"), 300, 48*time.Hour)
	writeFile(t, filepath.Join(root, "b.deb"), 200, 48*time.Hour)
	env := testEnv()
	env.LookPath = func(string) (string, error) { return "/usr/bin/apt-get", nil }
	var ran []string
	env.Run = func(_ context.Context, name string, args ...string) ([]byte, error) {
		ran = append([]string{name}, args...)
		return nil, os.Remove(filepath.Join(root, "a.deb"))
	}
	target := &Target{ID: "t", Roots: []string{root}, Command: []string{"apt-get", "clean"}}

	c := Clean(context.Background(), env, target)
	if strings.Join(ran, " ") != "apt-get clean" {
		t.Fatalf("ran %q", ran)
	}
	if c.Status != Done || c.Deleted != 1 || c.Freed != 300 {
		t.Fatalf("clean = %+v, want 1 file and 300 bytes", c)
	}

	env.LookPath = func(string) (string, error) { return "", errors.New("missing") }
	if s := Scan(context.Background(), env, target); s.Status != NotFound {
		t.Errorf("uninstalled command: status %s, want not_found", s.Status)
	}
}

func TestSelectNarrowsCandidates(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a"), 1, 48*time.Hour)
	writeFile(t, filepath.Join(root, "b"), 1, 48*time.Hour)
	target := &Target{ID: "t", Roots: []string{root}, Select: func(c []Candidate) []Candidate { return c[:1] }}
	if c := Clean(context.Background(), testEnv(), target); c.Deleted != 1 || !exists(filepath.Join(root, "b")) {
		t.Fatalf("clean = %+v", c)
	}
}

func TestFormatBytes(t *testing.T) {
	for n, want := range map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KB", 1536: "1.5 KB", 5 << 20: "5.0 MB", 3 << 30: "3.0 GB"} {
		if got := FormatBytes(n); got != want {
			t.Errorf("FormatBytes(%d) = %q, want %q", n, got, want)
		}
	}
}
