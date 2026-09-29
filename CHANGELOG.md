# Changelog

All notable changes to sweepkit are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-09-29

First public release.

### Added

- Command-line tool `sweepkit` with `list`, `scan`, `clean` and `version`.
  `scan` never changes anything; `clean` asks before deleting and supports
  `--yes` and `--json` for scripts.
- Desktop app `sweepkit-gui` in English and Turkish, following the system's
  light or dark theme.
- Windows targets: temporary files, system temporary files, Windows Update
  downloads, crash dumps, error reports and GPU shader caches.
- Linux targets: thumbnail cache, APT, DNF and pacman package caches, archived
  systemd journals and core dumps.
- Browser caches (Chrome, Chromium, Edge, Brave, Vivaldi, Opera, Firefox) and
  developer caches (npm, pip, Go, Yarn, NuGet) as opt-in targets.
- Safety model: age thresholds, sealed personal folders, protected system
  folders, no link or junction following, traversal-resistant deletion,
  in-use files left alone, browsers must be closed first.
- Release builds for Windows and Linux (amd64 and arm64) with SHA-256
  checksums and build provenance attestations.

[0.1.0]: https://github.com/bugraskl/sweepkit/releases/tag/v0.1.0
