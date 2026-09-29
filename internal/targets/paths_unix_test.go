//go:build !windows

package targets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserDirsReadsLocalizedFolders(t *testing.T) {
	file := filepath.Join(t.TempDir(), "user-dirs.dirs")
	content := `# written by xdg-user-dirs-update
XDG_DESKTOP_DIR="$HOME/Masaüstü"
XDG_DOCUMENTS_DIR="$HOME/Belgeler"
XDG_PUBLICSHARE_DIR="$HOME/"
XDG_MUSIC_DIR="/data/Müzik"
`
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := userDirs("/home/ayse", file)
	want := []string{"/home/ayse/Masaüstü", "/home/ayse/Belgeler", "/data/Müzik"}
	if len(got) != len(want) {
		t.Fatalf("userDirs = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("userDirs = %q, want %q", got, want)
		}
	}
}
