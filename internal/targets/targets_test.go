package targets

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bugraskl/sweepkit/internal/core"
)

func TestTargetDefinitions(t *testing.T) {
	for _, osName := range []string{"windows", "linux"} {
		p := Paths{OS: osName, Home: "/h", LocalAppData: "/h/l", RoamingAppData: "/h/r", SystemRoot: "/w", ProgramData: "/pd", CacheHome: "/h/.cache", ConfigHome: "/h/.config", DataHome: "/h/.local/share", GoCache: "/h/go-build"}
		seen := map[string]bool{}
		ts := For(p)
		if len(ts) == 0 {
			t.Fatalf("%s: no targets", osName)
		}
		for _, tg := range ts {
			if seen[tg.ID] {
				t.Errorf("%s: duplicate target ID %q", osName, tg.ID)
			}
			seen[tg.ID] = true
			if tg.Name.EN == "" || tg.Name.TR == "" || tg.Description.EN == "" || tg.Description.TR == "" {
				t.Errorf("%s/%s: missing name or description translation", osName, tg.ID)
			}
			if (tg.Caveat.EN == "") != (tg.Caveat.TR == "") {
				t.Errorf("%s/%s: caveat is only translated in one language", osName, tg.ID)
			}
			if tg.Category == core.Browser && len(tg.BlockedBy) == 0 {
				t.Errorf("%s/%s: browser caches must wait for the browser to close", osName, tg.ID)
			}
			if tg.Category == core.Browser || tg.Category == core.Developer || tg.Category == core.Graphics {
				if tg.Default {
					t.Errorf("%s/%s: caches that cost speed must be opt-in", osName, tg.ID)
				}
			}
		}
	}
	if For(Paths{OS: "plan9"}) != nil {
		t.Error("unsupported systems should get no targets")
	}
}

func TestPackageKey(t *testing.T) {
	cases := map[string]string{
		"linux-6.10.1.arch1-1-x86_64.pkg.tar.zst":        "linux x86_64",
		"lib32-mesa-24.1.3-1-x86_64.pkg.tar.zst":         "lib32-mesa x86_64",
		"python-pip-24.0-1-any.pkg.tar.zst":              "python-pip any",
		"xorg-server-common-21.1.13-1-x86_64.pkg.tar.xz": "xorg-server-common x86_64",
	}
	for file, want := range cases {
		if got, ok := packageKey(file); !ok || got != want {
			t.Errorf("packageKey(%q) = %q, %v; want %q", file, got, ok, want)
		}
	}
	for _, bad := range []string{"README", "foo.pkg.tar.zst", "a-b.pkg.tar.zst"} {
		if _, ok := packageKey(bad); ok {
			t.Errorf("packageKey(%q) accepted a malformed name", bad)
		}
	}
}

func TestKeepNewestKeepsTwoVersionsAndTheirSignatures(t *testing.T) {
	now := time.Now()
	c := func(rel string, age int) core.Candidate {
		return core.Candidate{Root: "/cache", Rel: rel, ModTime: now.Add(-time.Duration(age) * time.Hour)}
	}
	found := []core.Candidate{
		c("vim-9.1-1-x86_64.pkg.tar.zst", 30),
		c("vim-9.1-1-x86_64.pkg.tar.zst.sig", 30),
		c("vim-9.2-1-x86_64.pkg.tar.zst", 20),
		c("vim-9.3-1-x86_64.pkg.tar.zst", 10),
		c("gvim-9.0-1-x86_64.pkg.tar.zst", 40),
		c("strange-file.pkg.tar.zst", 50),
	}
	got := keepNewest(2)(found)
	var rels []string
	for _, g := range got {
		rels = append(rels, g.Rel)
	}
	want := []string{"vim-9.1-1-x86_64.pkg.tar.zst", "vim-9.1-1-x86_64.pkg.tar.zst.sig"}
	if len(rels) != len(want) || rels[0] != want[0] || rels[1] != want[1] {
		t.Fatalf("keepNewest removed %q, want %q", rels, want)
	}
}

func TestGoCacheLayoutGuardsAgainstMisconfiguration(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-48 * time.Hour)
	for _, rel := range []string{"README", "trim.txt", "ab/abc-a", "0f/0ff-d", "notes/todo.txt"} {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	var goCache *core.Target
	for _, tg := range developerTargets(Paths{OS: "linux", GoCache: dir}) {
		if tg.ID == "go-build-cache" {
			goCache = tg
		}
	}
	env := &core.Env{
		Now:      time.Now,
		Running:  func([]string) []string { return nil },
		LookPath: func(string) (string, error) { return "", errors.New("n/a") },
		Guard:    &core.Guard{},
	}
	r := core.Clean(context.Background(), env, goCache)
	if r.Deleted != 2 {
		t.Fatalf("deleted %d files, want the 2 cache entries", r.Deleted)
	}
	for _, keep := range []string{"README", "trim.txt", "notes/todo.txt"} {
		if _, err := os.Stat(filepath.Join(dir, keep)); err != nil {
			t.Errorf("%s was deleted", keep)
		}
	}
}

func TestGuardSealsPersonalFolders(t *testing.T) {
	home := t.TempDir()
	docs := filepath.Join(home, "Documents")
	cache := filepath.Join(home, "cache")
	for _, d := range []string{docs, cache} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	g := GuardFor(Paths{OS: "linux", Home: home, Personal: []string{docs}, CacheHome: cache})
	if _, err := g.CheckRoot(docs); !errors.Is(err, core.ErrRefused) {
		t.Errorf("Documents accepted as a root: %v", err)
	}
	if _, err := g.CheckRoot(home); !errors.Is(err, core.ErrRefused) {
		t.Errorf("home accepted as a root: %v", err)
	}
	if _, err := g.CheckRoot(cache); !errors.Is(err, core.ErrRefused) {
		t.Errorf("the whole cache folder accepted as a root: %v", err)
	}
	sub := filepath.Join(cache, "thumbnails")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CheckRoot(sub); err != nil {
		t.Errorf("a folder inside the cache was refused: %v", err)
	}
}
