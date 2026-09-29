package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bugraskl/sweepkit/internal/core"
)

type fixture struct {
	cli       *cli
	out, err  *bytes.Buffer
	junk, opt string
}

func newFixture(t *testing.T, interactive bool, stdin string) *fixture {
	t.Helper()
	dir := t.TempDir()
	junk := filepath.Join(dir, "temp", "junk.tmp")
	opt := filepath.Join(dir, "cache", "entry")
	old := time.Now().Add(-48 * time.Hour)
	for _, f := range []string{junk, opt} {
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, make([]byte, 2048), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(f, old, old); err != nil {
			t.Fatal(err)
		}
	}
	env := &core.Env{
		Now:      time.Now,
		Running:  func([]string) []string { return nil },
		LookPath: func(string) (string, error) { return "", errors.New("n/a") },
		Guard:    &core.Guard{},
	}
	ts := []*core.Target{
		{ID: "temp", Name: core.Text{EN: "Temporary files"}, Default: true, Roots: []string{filepath.Dir(junk)}},
		{ID: "cache", Name: core.Text{EN: "Some cache"}, Roots: []string{filepath.Dir(opt)}},
	}
	f := &fixture{out: &bytes.Buffer{}, err: &bytes.Buffer{}, junk: junk, opt: opt}
	f.cli = &cli{env: env, targets: ts, in: strings.NewReader(stdin), out: f.out, errOut: f.err, interactive: interactive}
	return f
}

func (f *fixture) run(args ...string) int { return f.cli.run(context.Background(), args) }

func present(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestScanNeverDeletes(t *testing.T) {
	f := newFixture(t, false, "")
	if code := f.run("scan", "--all"); code != 0 {
		t.Fatalf("exit %d: %s", code, f.err)
	}
	if !present(f.junk) || !present(f.opt) {
		t.Fatal("scan deleted files")
	}
	if !strings.Contains(f.out.String(), "4.0 KB in 2 files") {
		t.Errorf("unexpected output:\n%s", f.out)
	}
}

func TestCleanRefusesWithoutConfirmationWhenNotInteractive(t *testing.T) {
	f := newFixture(t, false, "y\n")
	if code := f.run("clean"); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !present(f.junk) {
		t.Fatal("deleted without --yes")
	}
}

func TestCleanAsksFirst(t *testing.T) {
	f := newFixture(t, true, "n\n")
	f.run("clean")
	if !present(f.junk) {
		t.Fatal("deleted although the answer was no")
	}

	f = newFixture(t, true, "y\n")
	if code := f.run("clean"); code != 0 {
		t.Fatalf("exit %d: %s", code, f.err)
	}
	if present(f.junk) {
		t.Error("the default target was not cleaned")
	}
	if !present(f.opt) {
		t.Error("an optional target was cleaned without --all")
	}
}

func TestCleanSelection(t *testing.T) {
	f := newFixture(t, false, "")
	if code := f.run("clean", "--yes", "--only", "cache"); code != 0 {
		t.Fatalf("exit %d: %s", code, f.err)
	}
	if present(f.opt) || !present(f.junk) {
		t.Fatal("--only did not limit the cleaning")
	}

	f = newFixture(t, false, "")
	if code := f.run("clean", "--yes", "--all", "--skip", "temp"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if present(f.opt) || !present(f.junk) {
		t.Fatal("--skip was not honoured")
	}

	if code := newFixture(t, false, "").run("scan", "--only", "nope"); code != 2 {
		t.Errorf("unknown target: exit %d, want 2", code)
	}
}

func TestCleanJSON(t *testing.T) {
	f := newFixture(t, false, "")
	if code := f.run("clean", "--yes", "--json"); code != 0 {
		t.Fatalf("exit %d: %s", code, f.err)
	}
	var results []core.CleanResult
	if err := json.Unmarshal(f.out.Bytes(), &results); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, f.out)
	}
	if len(results) != 1 || results[0].ID != "temp" || results[0].Deleted != 1 || results[0].Freed != 2048 {
		t.Fatalf("results = %+v", results)
	}
}
