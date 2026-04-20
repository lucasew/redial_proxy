* This software is meant to run in the same machine that is going to consume it. It only listens in the loopback interface.

## Operational Memory

* `cmd/redial_proxy/main.go` -> Application entry point, flag parsing, listener configuration, and proxy server initialization.
* `internal/dialer/` -> Custom `Redialer` implementation wrapping `net.Dialer` with backoff and retry logic.
