package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/bugraskl/sweepkit/internal/core"
	"github.com/bugraskl/sweepkit/internal/targets"
	"github.com/bugraskl/sweepkit/internal/version"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const projectURL = "https://github.com/bugraskl/sweepkit"

// App is the backend the desktop window talks to.
type App struct {
	ctx     context.Context
	env     *core.Env
	targets []*core.Target
	byID    map[string]*core.Target
	// busy serialises scans and cleans.
	busy sync.Mutex
}

// NewApp loads the targets for this computer.
func NewApp() *App {
	env, ts := targets.System()
	a := &App{env: env, targets: ts, byID: map[string]*core.Target{}}
	for _, t := range ts {
		a.byID[t.ID] = t
	}
	return a
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// TargetInfo describes a target to the window.
type TargetInfo struct {
	ID          string        `json:"id"`
	Category    core.Category `json:"category"`
	Name        core.Text     `json:"name"`
	Description core.Text     `json:"description"`
	Caveat      core.Text     `json:"caveat"`
	Admin       bool          `json:"admin"`
	Default     bool          `json:"default"`
}

// Info is everything the window needs to draw itself.
type Info struct {
	Version    string          `json:"version"`
	OS         string          `json:"os"`
	Elevated   bool            `json:"elevated"`
	CanElevate bool            `json:"canElevate"`
	Categories []core.Category `json:"categories"`
	Targets    []TargetInfo    `json:"targets"`
}

// Info describes the app, this system and every target.
func (a *App) Info() Info {
	info := Info{
		Version:    version.String(),
		OS:         runtime.GOOS,
		Elevated:   a.env.Elevated,
		CanElevate: canElevate && !a.env.Elevated,
		Categories: core.Categories,
	}
	for _, t := range a.targets {
		info.Targets = append(info.Targets, TargetInfo{
			ID: t.ID, Category: t.Category, Name: t.Name, Description: t.Description,
			Caveat: t.Caveat, Admin: t.Admin, Default: t.Default,
		})
	}
	return info
}

// Progress is sent to the window as the "progress" event.
type Progress struct {
	Phase string `json:"phase"`
	ID    string `json:"id"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
	// Result is set while scanning, so the list can fill in as it goes.
	Result *core.ScanResult `json:"result,omitempty"`
}

// Scan reports what cleaning each target would remove. It changes nothing.
func (a *App) Scan(ids []string) ([]core.ScanResult, error) {
	picked, err := a.pick(ids)
	if err != nil {
		return nil, err
	}
	if !a.busy.TryLock() {
		return nil, errors.New("busy")
	}
	defer a.busy.Unlock()

	results := make([]core.ScanResult, len(picked))
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	// Targets are independent, so scan a few at once.
	sem := make(chan struct{}, 4)
	for i, t := range picked {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := core.Scan(a.ctx, a.env, t)
			results[i] = r
			mu.Lock()
			done++
			a.emit(Progress{Phase: "scan", ID: t.ID, Done: done, Total: len(picked), Result: &r})
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results, nil
}

// Clean removes what each target's scan reports, one target at a time.
func (a *App) Clean(ids []string) ([]core.CleanResult, error) {
	picked, err := a.pick(ids)
	if err != nil {
		return nil, err
	}
	if !a.busy.TryLock() {
		return nil, errors.New("busy")
	}
	defer a.busy.Unlock()

	var results []core.CleanResult
	for i, t := range picked {
		a.emit(Progress{Phase: "clean", ID: t.ID, Done: i, Total: len(picked)})
		results = append(results, core.Clean(a.ctx, a.env, t))
	}
	a.emit(Progress{Phase: "clean", Done: len(picked), Total: len(picked)})
	return results, nil
}

// RestartAsAdmin starts a new elevated copy of the app and closes this one.
func (a *App) RestartAsAdmin() error {
	if err := relaunchElevated(); err != nil {
		return err
	}
	wruntime.Quit(a.ctx)
	return nil
}

// OpenProjectPage opens the source code in the default browser.
func (a *App) OpenProjectPage() { wruntime.BrowserOpenURL(a.ctx, projectURL) }

func (a *App) pick(ids []string) ([]*core.Target, error) {
	var picked []*core.Target
	for _, id := range ids {
		t := a.byID[id]
		if t == nil {
			return nil, fmt.Errorf("unknown target %q", id)
		}
		picked = append(picked, t)
	}
	return picked, nil
}

func (a *App) emit(p Progress) {
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, "progress", p)
	}
}
