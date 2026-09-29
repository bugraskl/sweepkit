package targets

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Only the standard syscall package is used here, so the command-line tool
// does not link any networking code (golang.org/x/sys/windows would).
var (
	procSHGetKnownFolderPath = systemDLL("shell32.dll").NewProc("SHGetKnownFolderPath")
	procCoTaskMemFree        = systemDLL("ole32.dll").NewProc("CoTaskMemFree")
)

// systemDLL loads a DLL from the System32 folder by its full path, so a
// planted copy next to sweepkit.exe is never used. kernel32.dll is already
// loaded by the Go runtime and is safe to name directly.
func systemDLL(name string) *syscall.LazyDLL {
	buf := make([]uint16, syscall.MAX_PATH)
	n, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemDirectoryW").Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 || int(n) >= len(buf) {
		// Fall back to a name that cannot load anything.
		return syscall.NewLazyDLL(`\\?\unavailable\` + name)
	}
	return syscall.NewLazyDLL(filepath.Join(syscall.UTF16ToString(buf[:n]), name))
}

// knownFolderID is a KNOWNFOLDERID; the values are from ShlObj_core.h.
type knownFolderID struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

var (
	folderDocuments       = knownFolderID{0xfdd39ad0, 0x238f, 0x46af, [8]byte{0xad, 0xb4, 0x6c, 0x85, 0x48, 0x03, 0x69, 0xc7}}
	folderDesktop         = knownFolderID{0xb4bfcc3a, 0xdb2c, 0x424c, [8]byte{0xb0, 0x29, 0x7f, 0xe9, 0x9a, 0x87, 0xc6, 0x41}}
	folderDownloads       = knownFolderID{0x374de290, 0x123f, 0x4565, [8]byte{0x91, 0x64, 0x39, 0xc4, 0x92, 0x5e, 0x46, 0x7b}}
	folderPictures        = knownFolderID{0x33e28130, 0x4e1e, 0x4676, [8]byte{0x83, 0x5a, 0x98, 0x39, 0x5c, 0x3b, 0xc3, 0xbb}}
	folderMusic           = knownFolderID{0x4bd8d571, 0x6d19, 0x48d3, [8]byte{0xbe, 0x97, 0x42, 0x22, 0x20, 0x08, 0x0e, 0x43}}
	folderVideos          = knownFolderID{0x18989b1d, 0x99b5, 0x455b, [8]byte{0x84, 0x1c, 0xab, 0x7c, 0x74, 0xe4, 0xdd, 0xfc}}
	folderSavedGames      = knownFolderID{0x4c5c32ff, 0xbb9d, 0x43b0, [8]byte{0xb5, 0xb4, 0x2d, 0x72, 0xe5, 0x4e, 0xaa, 0xa4}}
	folderFavorites       = knownFolderID{0x1777f761, 0x68ad, 0x4d8a, [8]byte{0x87, 0xbd, 0x30, 0xb7, 0x59, 0xfa, 0x33, 0xdd}}
	folderContacts        = knownFolderID{0x56784854, 0xc6cb, 0x462b, [8]byte{0x81, 0x69, 0x88, 0xe3, 0x50, 0xac, 0xb8, 0x82}}
	folderLocalAppDataLow = knownFolderID{0xa520a1a4, 0x1780, 0x4ff6, [8]byte{0xbd, 0x18, 0x16, 0x73, 0x43, 0xc5, 0xaf, 0x16}}
)

func detect() Paths {
	home := firstAbs(os.Getenv("USERPROFILE"), mustHome())
	local := firstAbs(os.Getenv("LOCALAPPDATA"), join(home, "AppData", "Local"))
	p := Paths{
		Home:            home,
		LocalAppData:    local,
		RoamingAppData:  firstAbs(os.Getenv("APPDATA"), join(home, "AppData", "Roaming")),
		LocalLow:        firstAbs(knownFolder(&folderLocalAppDataLow), join(home, "AppData", "LocalLow")),
		SystemRoot:      firstAbs(os.Getenv("SystemRoot"), `C:\Windows`),
		ProgramData:     firstAbs(os.Getenv("ProgramData"), `C:\ProgramData`),
		ProgramFiles:    firstAbs(os.Getenv("ProgramFiles"), `C:\Program Files`),
		ProgramFilesX86: firstAbs(os.Getenv("ProgramFiles(x86)"), `C:\Program Files (x86)`),
		GoCache:         firstAbs(os.Getenv("GOCACHE"), goEnvCache(), join(local, "go-build")),
		Getenv:          os.Getenv,
	}
	p.UserTemp = tempDirs(os.Getenv("TEMP"), os.Getenv("TMP"), join(local, "Temp"))

	for _, id := range []*knownFolderID{
		&folderDocuments, &folderDesktop, &folderDownloads, &folderPictures, &folderMusic,
		&folderVideos, &folderSavedGames, &folderFavorites, &folderContacts,
	} {
		if dir := knownFolder(id); dir != "" {
			p.Personal = append(p.Personal, dir)
		}
	}
	for _, name := range []string{"Documents", "Desktop", "Downloads", "Pictures", "Music", "Videos", "Saved Games", "Favorites", "Contacts", "OneDrive"} {
		p.Personal = append(p.Personal, join(home, name))
	}
	for _, key := range []string{"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
		if dir := firstAbs(os.Getenv(key)); dir != "" {
			p.Personal = append(p.Personal, dir)
		}
	}
	return p
}

// knownFolder returns where Windows keeps a known folder, which may have
// been moved to another drive or into OneDrive.
func knownFolder(id *knownFolderID) string {
	if procSHGetKnownFolderPath.Find() != nil || procCoTaskMemFree.Find() != nil {
		return ""
	}
	var p *uint16
	hr, _, _ := procSHGetKnownFolderPath.Call(uintptr(unsafe.Pointer(id)), 0, 0, uintptr(unsafe.Pointer(&p)))
	if p != nil {
		defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(p)))
	}
	if hr != 0 || p == nil {
		return ""
	}
	return syscall.UTF16ToString(unsafe.Slice(p, pathLen(p)))
}

func pathLen(p *uint16) int {
	n := 0
	for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; ptr = unsafe.Add(ptr, 2) {
		n++
	}
	return n
}

// tempDirs keeps the candidate temp folders that really look like temp
// folders, so a TEMP variable pointing somewhere odd is ignored.
func tempDirs(candidates ...string) []string {
	var out []string
	for _, c := range candidates {
		c = firstAbs(c)
		if c == "" || !looksLikeTemp(c) {
			continue
		}
		dup := false
		for _, o := range out {
			if strings.EqualFold(o, c) {
				dup = true
			}
		}
		if !dup {
			out = append(out, c)
		}
	}
	return out
}

// looksLikeTemp reports whether one of the path's folders is named Temp or
// Tmp. Windows puts per-session temp folders below Temp, e.g. Temp\2.
func looksLikeTemp(path string) bool {
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if strings.EqualFold(part, "temp") || strings.EqualFold(part, "tmp") {
			return true
		}
	}
	return false
}

func mustHome() string {
	home, _ := os.UserHomeDir()
	return home
}
