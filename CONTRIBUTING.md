# Contributing

Thanks for helping. sweepkit's value comes from being boring and trustworthy,
so the bar for new cleaning targets is high.

## Adding a target

A new target needs all of these:

1. **A real, measurable benefit.** It must free disk space that stays freed
   long enough to matter. "Speeds up your PC" is not a benefit unless it can
   be measured and reproduced.
2. **A documented or well-established location.** Link to the vendor's
   documentation or the tool's own clean command in the pull request.
3. **Safe defaults.** Pick a `MinAge` that leaves files in active use alone.
   Anything that costs speed (browser, shader and developer caches) must be
   opt-in (`Default: false`) and explain the cost in `Caveat`.
4. **Both languages.** Every `Name`, `Description` and `Caveat` needs English
   and Turkish text.
5. **Tests** when the target has any logic of its own (`Match`, `Select`).

If a tool has its own clean command, prefer running it (`Command`) over
deleting its files yourself.

## Development

You need Go 1.26 or newer. The desktop app also needs
[Wails v2](https://wails.io/docs/gettingstarted/installation) and, on Linux,
`libgtk-3-dev` and `libwebkit2gtk-4.1-dev`.

```sh
go test ./internal/... ./cmd/...          # tests
go run ./cmd/sweepkit scan --all          # try the CLI; scan changes nothing
wails dev                                 # desktop app with live reload
wails build -tags webkit2_41              # desktop app build on Linux
```

The front end is plain HTML, CSS and JavaScript in `frontend/dist`; there is
no build step. The icon is drawn by `go run scripts/genicon/main.go`.

## Releasing

1. Update `CHANGELOG.md`.
2. Tag and push: `git tag v0.2.0 && git push origin v0.2.0`.

The release workflow builds everything, attests it and publishes the release.
