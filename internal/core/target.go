// Package core finds reclaimable files and removes them without ever
// stepping outside the directories a target names.
package core

import (
	"context"
	"path/filepath"
	"time"
)

// Text is a user-facing string in every language the app ships.
type Text struct {
	EN string `json:"en"`
	TR string `json:"tr"`
}

// Category groups targets in listings.
type Category string

const (
	System    Category = "system"
	Logs      Category = "logs"
	Packages  Category = "packages"
	Graphics  Category = "graphics"
	Browser   Category = "browser"
	Developer Category = "developer"
)

// Categories lists every category in display order.
var Categories = []Category{System, Logs, Packages, Graphics, Browser, Developer}

// Target describes one kind of reclaimable data and where it lives.
type Target struct {
	ID          string
	Category    Category
	Name        Text
	Description Text
	// Caveat explains what cleaning costs, if anything.
	Caveat Text
	// Admin marks targets that need elevated privileges to clean.
	Admin bool
	// Default marks targets that are selected unless the user says otherwise.
	Default bool
	// MinAge keeps files that were modified more recently than this.
	MinAge time.Duration

	// Roots are directories whose contents may be removed. The directories
	// themselves are always kept.
	Roots []string
	// Files are individual files that may be removed.
	Files []string
	// Match, if set, limits which files under Roots qualify. rel is relative
	// to the root and uses forward slashes.
	Match func(rel string) bool
	// BlockedBy lists process names that must not be running while cleaning.
	BlockedBy []string
	// Select, if set, narrows the files found under Roots, e.g. to keep the
	// newest versions of each package.
	Select func(found []Candidate) []Candidate
	// Command, if set, cleans by running a tool instead of deleting files one
	// by one. Candidates still estimate the size and measure what was freed.
	Command []string
}

// Candidate is a file that cleaning a target would remove.
type Candidate struct {
	// Root is the directory the file was found under. For Target.Files it
	// is the file's parent directory.
	Root    string
	Rel     string
	Size    int64
	ModTime time.Time
}

// Path returns the candidate's absolute path.
func (c Candidate) Path() string { return filepath.Join(c.Root, c.Rel) }

// Env is everything the engine needs from the running system. Tests swap
// in fakes.
type Env struct {
	Elevated bool
	Now      func() time.Time
	// Running returns the subset of names that belong to running processes.
	Running func(names []string) []string
	// LookPath reports where a command is installed.
	LookPath func(name string) (string, error)
	// Run runs a command and returns its combined output.
	Run func(ctx context.Context, name string, args ...string) ([]byte, error)
	// Guard decides which directories may be cleaned at all.
	Guard *Guard
}

// Status is the outcome of scanning or cleaning a target.
type Status string

const (
	// Ready means the target was scanned and can be cleaned.
	Ready Status = "ready"
	// NotFound means none of the target's locations exist on this system.
	NotFound Status = "not_found"
	// NeedsAdmin means cleaning requires elevated privileges.
	NeedsAdmin Status = "needs_admin"
	// Blocked means a program that uses the files is running.
	Blocked Status = "blocked"
	// Failed means the target could not be scanned or cleaned.
	Failed Status = "failed"
	// Done means cleaning ran.
	Done Status = "done"
)

// ScanResult reports what cleaning a target would remove.
type ScanResult struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Files  int    `json:"files"`
	Bytes  int64  `json:"bytes"`
	// Partial means some folders could not be read, so Files and Bytes are
	// a lower bound.
	Partial  bool     `json:"partial,omitempty"`
	Blockers []string `json:"blockers,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// CleanResult reports what cleaning a target removed.
type CleanResult struct {
	ID      string `json:"id"`
	Status  Status `json:"status"`
	Deleted int    `json:"deleted"`
	Freed   int64  `json:"freed"`
	// InUse counts files another program had open, which were left alone.
	InUse int `json:"inUse,omitempty"`
	// Failed counts files that could not be deleted for other reasons.
	Failed   int      `json:"failed,omitempty"`
	Blockers []string `json:"blockers,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Error    string   `json:"error,omitempty"`
}
