## IGNORE: SOCKS5 Authentication

**- Pattern:** Implementing Username/Password authentication for the SOCKS5 proxy.
**- Justification:** The proxy is intended for local-only use on the loopback interface (`127.0.0.1`) as per `AGENTS.md`. Adding authentication introduces unnecessary complexity and overhead for this specific use case.
**- Files Affected:** `cmd/redial_proxy/main.go`, `go.mod`, `go.sum`, `internal/dialer/*`

## IGNORE: Committing Vendor Directory

**- Pattern:** Committing the `vendor/` directory to the repository.
**- Justification:** Go modules handle dependencies. Committing `vendor/` bloats the repository and is generally discouraged unless specifically required.
**- Files Affected:** `vendor/`

## IGNORE: Downgrading GitHub Actions

**- Pattern:** Downgrading `actions/checkout` (e.g., to v4) or `mise-action` (e.g., to v2).
**- Justification:** Global instructions explicitly forbid downgrading these dependencies. Always use the latest stable versions as per project guidelines.
**- Files Affected:** `.github/workflows/*.yml`

## IGNORE: Unpinning Docker Base Images

**- Pattern:** Replacing specific SHA digests in Dockerfiles with mutable tags (e.g., `golang:1.25.6` instead of `golang:1.25@sha256:...`).
**- Justification:** Pinned digests ensure reproducibility and security. Using mutable tags introduces risk. And there is Renovate to suggest dependency updates already.
**- Files Affected:** `Dockerfile`

## IGNORE: Misconfigured golangci-lint

**- Pattern:** Removing `golangci-lint` from `[tools]` in `mise.toml` or attempting to install it via `go install` in tasks.
**- Justification:** `golangci-lint` must be managed via `mise` tools and pinned to a specific version to ensure stability and proper integration with the development workflow.
**- Files Affected:** `mise.toml`

## IGNORE: Weakening CI Checks

**- Pattern:** Removing critical steps (like `build`) from the `ci` task in `mise.toml`.
**- Justification:** The `ci` task is explicitly defined to run `lint`, `test`, and `build`. Removing any of these steps compromises the verification process.
**- Files Affected:** `mise.toml`
