package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilesInUseAreLeftAlone(t *testing.T) {
	root := t.TempDir()
	busy := filepath.Join(root, "busy.tmp")
	writeFile(t, busy, 10, 48*time.Hour)
	writeFile(t, filepath.Join(root, "idle.tmp"), 10, 48*time.Hour)

	// Go opens files without FILE_SHARE_DELETE, like most Windows programs.
	f, err := os.Open(busy)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	c := Clean(context.Background(), testEnv(), &Target{ID: "t", Roots: []string{root}})
	if c.InUse != 1 || c.Deleted != 1 || c.Failed != 0 {
		t.Fatalf("clean = %+v, want 1 deleted and 1 in use", c)
	}
	if !exists(busy) {
		t.Error("an open file was deleted")
	}
}

func TestReadOnlyFilesAreDeleted(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "readonly.tmp")
	writeFile(t, file, 10, 48*time.Hour)
	if err := os.Chmod(file, 0o444); err != nil {
		t.Fatal(err)
	}
	if c := Clean(context.Background(), testEnv(), &Target{ID: "t", Roots: []string{root}}); c.Deleted != 1 || exists(file) {
		t.Fatalf("clean = %+v", c)
	}
}
