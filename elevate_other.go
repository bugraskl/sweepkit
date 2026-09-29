//go:build !windows

package main

import "errors"

// Restarting a desktop app as root is not a good idea on Linux; the window
// points to `sudo sweepkit clean` instead.
const canElevate = false

func relaunchElevated() error {
	return errors.New("run `sudo sweepkit clean` in a terminal to clean system files")
}
