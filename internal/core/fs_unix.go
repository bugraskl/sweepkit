//go:build !windows

package core

import (
	"io/fs"
	"path/filepath"
	"syscall"
)

// resolve returns the real location of path with every symlink followed.
func resolve(path string) (string, error) { return filepath.EvalSymlinks(path) }

// isLink reports whether info describes a symlink or other non-regular
// entry that must never be followed or deleted.
func isLink(info fs.FileInfo) bool {
	return info.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0
}

// Unix lets programs delete files that are open elsewhere, so nothing is
// ever reported as in use.
func isInUse(error) bool { return false }

// deviceID returns the filesystem a file lives on, so a walk never crosses
// into another mount.
func deviceID(info fs.FileInfo) (uint64, bool) {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(st.Dev), true // Dev's type differs between platforms.
	}
	return 0, false
}
