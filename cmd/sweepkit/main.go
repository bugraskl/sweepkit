// Command sweepkit shows and removes files that are safe to delete.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/bugraskl/sweepkit/internal/core"
	"github.com/bugraskl/sweepkit/internal/targets"
	"github.com/bugraskl/sweepkit/internal/version"
)

const usage = `sweepkit - honest disk cleanup for Windows and Linux

Usage:
  sweepkit list  [--json]                     show every target
  sweepkit scan  [selection] [--json]         show what would be deleted
  sweepkit clean [selection] [--yes] [--json] delete it
  sweepkit version

Selection (scan and clean act on the same targets):
  (none)          targets marked as default
  --all           every target, including optional ones
  --only a,b      exactly these targets
  --skip a,b      leave these out

scan never changes anything. clean asks before deleting unless --yes is given.
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	env, ts := targets.System()
	interactive := isTerminal(os.Stdin) && isTerminal(os.Stderr)
	c := &cli{env: env, targets: ts, in: os.Stdin, out: os.Stdout, errOut: os.Stderr, interactive: interactive}
	os.Exit(c.run(ctx, os.Args[1:]))
}

type cli struct {
	env         *core.Env
	targets     []*core.Target
	in          io.Reader
	out, errOut io.Writer
	interactive bool
}

func (c *cli) run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprint(c.errOut, usage)
		return 2
	}
	if len(c.targets) == 0 {
		fmt.Fprintf(c.errOut, "sweepkit does not support %s yet.\n", runtime.GOOS)
		return 1
	}
	switch args[0] {
	case "list":
		return c.list(args[1:])
	case "scan":
		return c.scan(ctx, args[1:])
	case "clean":
		return c.clean(ctx, args[1:])
	case "version", "--version", "-v":
		fmt.Fprintf(c.out, "sweepkit %s %s/%s\n", version.String(), runtime.GOOS, runtime.GOARCH)
		return 0
	case "help", "--help", "-h":
		fmt.Fprint(c.out, usage)
		return 0
	}
	fmt.Fprintf(c.errOut, "unknown command %q\n\n%s", args[0], usage)
	return 2
}

func (c *cli) list(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	asJSON := fs.Bool("json", false, "print JSON")
	if fs.Parse(args) != nil {
		return 2
	}
	type row struct {
		ID          string        `json:"id"`
		Category    core.Category `json:"category"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Default     bool          `json:"default"`
		Admin       bool          `json:"admin"`
		Present     bool          `json:"present"`
	}
	var rows []row
	for _, t := range c.targets {
		rows = append(rows, row{t.ID, t.Category, t.Name.EN, t.Description.EN, t.Default, t.Admin, core.Present(c.env, t)})
	}
	if *asJSON {
		return c.printJSON(rows)
	}
	w := tabwriter.NewWriter(c.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDEFAULT\tADMIN\tON THIS SYSTEM")
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.ID, r.Name, yesNo(r.Default), yesNo(r.Admin), yesNo(r.Present))
	}
	w.Flush()
	return 0
}

type selection struct {
	all        bool
	only, skip string
	asJSON     bool
}

func (s *selection) register(fs *flag.FlagSet) {
	fs.BoolVar(&s.all, "all", false, "include optional targets")
	fs.StringVar(&s.only, "only", "", "comma-separated target IDs to use instead of the defaults")
	fs.StringVar(&s.skip, "skip", "", "comma-separated target IDs to leave out")
	fs.BoolVar(&s.asJSON, "json", false, "print JSON")
}

// pick returns the selected targets and how many optional ones were left out.
func (c *cli) pick(s selection) ([]*core.Target, int, error) {
	byID := map[string]*core.Target{}
	for _, t := range c.targets {
		byID[t.ID] = t
	}
	skip := map[string]bool{}
	for _, id := range splitIDs(s.skip) {
		if byID[id] == nil {
			return nil, 0, fmt.Errorf("unknown target %q (see `sweepkit list`)", id)
		}
		skip[id] = true
	}
	var picked []*core.Target
	if s.only != "" {
		for _, id := range splitIDs(s.only) {
			t := byID[id]
			if t == nil {
				return nil, 0, fmt.Errorf("unknown target %q (see `sweepkit list`)", id)
			}
			if !skip[id] {
				picked = append(picked, t)
			}
		}
		return picked, 0, nil
	}
	optional := 0
	for _, t := range c.targets {
		switch {
		case skip[t.ID]:
		case t.Default || s.all:
			picked = append(picked, t)
		case core.Present(c.env, t):
			optional++
		}
	}
	return picked, optional, nil
}

func (c *cli) scan(ctx context.Context, args []string) int {
	var s selection
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	s.register(fs)
	if fs.Parse(args) != nil {
		return 2
	}
	picked, optional, err := c.pick(s)
	if err != nil {
		fmt.Fprintln(c.errOut, err)
		return 2
	}
	results := c.scanAll(ctx, picked, !s.asJSON)
	if s.asJSON {
		return c.printJSON(results)
	}
	c.printScan(picked, results, optional)
	if total, _ := readyTotal(results); total > 0 {
		fmt.Fprintln(c.out, "\nNothing has been deleted. Run `sweepkit clean` with the same options to delete these files.")
	}
	return 0
}

func (c *cli) clean(ctx context.Context, args []string) int {
	var s selection
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	s.register(fs)
	yes := fs.Bool("yes", false, "delete without asking")
	if fs.Parse(args) != nil {
		return 2
	}
	picked, optional, err := c.pick(s)
	if err != nil {
		fmt.Fprintln(c.errOut, err)
		return 2
	}
	if !*yes && (s.asJSON || !c.interactive) {
		fmt.Fprintln(c.errOut, "refusing to delete without confirmation: pass --yes when not running in a terminal")
		return 2
	}

	scans := c.scanAll(ctx, picked, !s.asJSON)
	if !s.asJSON {
		c.printScan(picked, scans, optional)
	}
	bytes, files := readyTotal(scans)
	if files == 0 {
		if s.asJSON {
			return c.printJSON([]core.CleanResult{})
		}
		fmt.Fprintln(c.out, "\nNothing to clean.")
		return 0
	}
	if !*yes {
		fmt.Fprintf(c.out, "\nPermanently delete %d files (%s)? They are not moved to the %s. [y/N] ", files, core.FormatBytes(bytes), trashName())
		answer, _ := bufio.NewReader(c.in).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			fmt.Fprintln(c.out, "Nothing was deleted.")
			return 1
		}
	}

	byID := map[string]*core.Target{}
	for _, t := range picked {
		byID[t.ID] = t
	}
	var results []core.CleanResult
	var freed int64
	failed := false
	if !s.asJSON {
		fmt.Fprintln(c.out)
	}
	for _, sr := range scans {
		if sr.Status != core.Ready || sr.Files == 0 {
			continue
		}
		t := byID[sr.ID]
		if !s.asJSON {
			fmt.Fprintf(c.out, "Cleaning %s... ", t.Name.EN)
		}
		r := core.Clean(ctx, c.env, t)
		results = append(results, r)
		freed += r.Freed
		failed = failed || r.Status == core.Failed
		if !s.asJSON {
			fmt.Fprintln(c.out, describeClean(r))
		}
		if ctx.Err() != nil {
			break
		}
	}
	if s.asJSON {
		c.printJSON(results)
	} else {
		fmt.Fprintf(c.out, "\nFreed %s.\n", core.FormatBytes(freed))
	}
	if failed || ctx.Err() != nil {
		return 1
	}
	return 0
}

func (c *cli) scanAll(ctx context.Context, picked []*core.Target, progress bool) []core.ScanResult {
	var results []core.ScanResult
	for i, t := range picked {
		if progress && c.interactive {
			fmt.Fprintf(c.errOut, "\rScanning %d/%d...", i+1, len(picked))
		}
		results = append(results, core.Scan(ctx, c.env, t))
	}
	if progress && c.interactive {
		fmt.Fprint(c.errOut, "\r                    \r")
	}
	return results
}

func (c *cli) printScan(picked []*core.Target, results []core.ScanResult, optional int) {
	fmt.Fprintf(c.out, "%10s  %7s  %s\n", "SIZE", "FILES", "TARGET")
	needsAdmin := false
	for i, r := range results {
		if r.Status == core.NotFound {
			continue
		}
		size, files := core.FormatBytes(r.Bytes), fmt.Sprint(r.Files)
		switch {
		case r.Partial && r.Files == 0:
			size, files = "-", "-" // unreadable without more rights
		case r.Partial:
			size += "+" // at least this much
		}
		note := ""
		switch r.Status {
		case core.NeedsAdmin:
			note, needsAdmin = " (needs administrator)", true
		case core.Blocked:
			note = fmt.Sprintf(" (close %s first)", strings.Join(r.Blockers, ", "))
		case core.Failed:
			note = " (error: " + r.Error + ")"
		}
		fmt.Fprintf(c.out, "%10s  %7s  %s [%s]%s\n", size, files, picked[i].Name.EN, r.ID, note)
	}
	bytes, files := readyTotal(results)
	fmt.Fprintf(c.out, "\nReady to clean: %s in %d files.\n", core.FormatBytes(bytes), files)
	if needsAdmin {
		if runtime.GOOS == "windows" {
			fmt.Fprintln(c.out, "Some targets need administrator rights: run sweepkit from a terminal opened with \"Run as administrator\".")
		} else {
			fmt.Fprintln(c.out, "Some targets need root: run sweepkit with sudo.")
		}
	}
	if optional > 0 {
		fmt.Fprintf(c.out, "%d optional targets were not included (browser and developer caches and more). Use --all or --only to include them.\n", optional)
	}
	for _, r := range results {
		for _, w := range r.Warnings {
			fmt.Fprintf(c.errOut, "warning: %s\n", w)
		}
	}
}

func describeClean(r core.CleanResult) string {
	switch r.Status {
	case core.Done:
		s := fmt.Sprintf("deleted %d files, freed %s", r.Deleted, core.FormatBytes(r.Freed))
		if r.InUse > 0 {
			s += fmt.Sprintf(", %d in use and left alone", r.InUse)
		}
		if r.Failed > 0 {
			s += fmt.Sprintf(", %d could not be deleted", r.Failed)
		}
		if len(r.Errors) > 0 {
			s += "\n  " + strings.Join(r.Errors, "\n  ")
		}
		if r.Error != "" {
			s += " (" + r.Error + ")"
		}
		return s
	case core.Blocked:
		return "skipped: close " + strings.Join(r.Blockers, ", ") + " first"
	case core.NeedsAdmin:
		return "skipped: needs administrator rights"
	case core.NotFound:
		return "nothing there any more"
	}
	return "failed: " + r.Error
}

func readyTotal(results []core.ScanResult) (bytes int64, files int) {
	for _, r := range results {
		if r.Status == core.Ready {
			bytes += r.Bytes
			files += r.Files
		}
	}
	return bytes, files
}

func (c *cli) printJSON(v any) int {
	enc := json.NewEncoder(c.out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		fmt.Fprintln(c.errOut, err)
		return 1
	}
	return 0
}

func splitIDs(s string) []string {
	var ids []string
	for _, id := range strings.Split(s, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "-"
}

func trashName() string {
	if runtime.GOOS == "windows" {
		return "Recycle Bin"
	}
	return "Trash"
}
