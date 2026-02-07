* This software is meant to run in the same machine that is going to consume it. It only listens in the loopback interface.

## Tooling

* This project uses `mise` for task management.
* Run `mise run ci` to run linters, tests, and build.
* Run `mise run build-all` to cross-compile for all platforms.
* Do not bypass `mise` tasks.
