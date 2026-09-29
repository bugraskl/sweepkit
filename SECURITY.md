# Security policy

sweepkit deletes files, so a bug that makes it delete the wrong thing is a
security bug. Examples:

- deleting anything outside the folders a target names,
- following a symlink, junction or other reparse point,
- touching a file inside a personal folder (Documents, Desktop, Downloads…),
- cleaning a target that needs administrator rights without them,
- any network access, or any data leaving the computer.

## Reporting

Please report these privately through
[GitHub security advisories](https://github.com/bugraskl/sweepkit/security/advisories/new)
rather than in a public issue. Include the operating system, the sweepkit
version (`sweepkit version`) and the steps or folder layout that trigger it.

You should get a reply within a week. Fixes are released as soon as they are
ready and credited in the changelog unless you prefer otherwise.

## Supported versions

Only the latest release receives fixes.
