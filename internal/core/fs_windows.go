package core

import (
	"errors"
	"io/fs"
	"strings"
	"syscall"
	"unsafe"
)

// Only the standard syscall package is used here, so the command-line tool
// does not link any networking code (golang.org/x/sys/windows would).
// kernel32.dll is already loaded by the Go runtime, so loading it by name
// cannot pick up a planted copy.
var procGetFinalPathNameByHandleW = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")

// resolve returns the real location of path, following symlinks, junctions
// and mount points and expanding short 8.3 names. filepath.EvalSymlinks
// does not follow junctions, so it cannot be used for this.
func resolve(path string) (string, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	share := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)
	h, err := syscall.CreateFile(p, 0, share, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", &fs.PathError{Op: "open", Path: path, Err: err}
	}
	defer syscall.CloseHandle(h)

	// Flags 0 means FILE_NAME_NORMALIZED | VOLUME_NAME_DOS.
	buf := make([]uint16, syscall.MAX_PATH)
	for {
		n, _, callErr := procGetFinalPathNameByHandleW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
		if n == 0 {
			return "", &fs.PathError{Op: "resolve", Path: path, Err: callErr}
		}
		if int(n) < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]uint16, n+1)
	}
	s := syscall.UTF16ToString(buf)
	switch {
	case strings.HasPrefix(s, `\\?\UNC\`):
		s = `\\` + s[len(`\\?\UNC\`):]
	case strings.HasPrefix(s, `\\?\`):
		s = s[len(`\\?\`):]
	}
	return s, nil
}

// isLink reports whether info describes a symlink, junction, mount point or
// any other reparse point, including cloud placeholders. None are followed
// or deleted.
func isLink(info fs.FileInfo) bool {
	if info.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 {
		return true
	}
	if a, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return a.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
	}
	return false
}

// Windows error codes returned when another process holds a file open.
const (
	errorSharingViolation syscall.Errno = 32
	errorLockViolation    syscall.Errno = 33
)

func isInUse(err error) bool {
	return errors.Is(err, errorSharingViolation) || errors.Is(err, errorLockViolation)
}

// deviceID is not needed on Windows: mount points are reparse points and
// are already skipped by isLink.
func deviceID(fs.FileInfo) (uint64, bool) { return 0, false }
