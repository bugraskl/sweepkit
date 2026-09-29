package targets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTempDirsIgnoresFoldersThatAreNotTemp(t *testing.T) {
	got := tempDirs(
		`C:\Users\me\AppData\Local\Temp`,
		`C:\Users\me\AppData\Local\temp`, // same folder, other case
		`C:\Users\me\Projects`,           // a TEMP variable pointing somewhere odd
		`C:\Users\me\AppData\Local\Temp\2`,
		`relative\Temp`,
		"",
	)
	want := []string{`C:\Users\me\AppData\Local\Temp`, `C:\Users\me\AppData\Local\Temp\2`}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("tempDirs = %q, want %q", got, want)
	}
}

func TestKnownFoldersAreFound(t *testing.T) {
	docs := knownFolder(&folderDocuments)
	if !filepath.IsAbs(docs) {
		t.Fatalf("Documents folder = %q, want an absolute path", docs)
	}
	if low := knownFolder(&folderLocalAppDataLow); !strings.HasSuffix(strings.ToLower(low), `\appdata\locallow`) {
		t.Errorf("LocalLow folder = %q", low)
	}
	sealed := false
	for _, p := range Detect().Personal {
		if strings.EqualFold(p, docs) {
			sealed = true
		}
	}
	if !sealed {
		t.Error("the real Documents folder is not in the sealed list")
	}
	if home, _ := os.UserHomeDir(); home == "" {
		t.Skip("no home folder")
	}
}
