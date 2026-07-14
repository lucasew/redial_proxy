# redial_proxy

Sometimes the internet here is a bit buggy — sites not loading CSS on the first
try, or the browser showing errors as if you were offline.

This is a small SOCKS5 proxy with a custom dialer that retries when the error
looks like a routing failure (message contains `route`), waiting 100ms between
attempts by default.

It is intended for local use only (loopback). See `AGENTS.md`.

## Layout

- `cmd/redial_proxy` — CLI entrypoint
- `internal/dialer` — retrying dialer
- `internal/errorreport` — fatal error helper

## Installing

```
mise use github:lucasew/redial_proxy
```

## Running

```
redial_proxy -h
```

Useful flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `-p` | `8889` | listen port |
| `-H` | `127.0.0.1` | listen host (keep loopback) |
| `-retries` | `3` | max dial retries on route-like errors |
| `-retry-delay` | `100ms` | delay between dial retries |

## Development

```
mise run ci   # lint, test, build
```
