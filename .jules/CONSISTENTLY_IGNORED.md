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

## IGNORE: Unintentional go.sum Modifications

**- Pattern:** Modifying `go.sum` to remove or alter unrelated dependencies during unrelated changes (e.g., refactoring or janitorial tasks).
**- Justification:** `go.sum` should only be modified intentionally, usually via `go mod tidy` or when explicitly adding/updating dependencies. Removing lines manually causes build/test failures.
**- Files Affected:** `go.sum`

## IGNORE: Premature Abstraction (Rule of Three)

**- Pattern:** Extracting simple logic (like flag parsing and environment variable setup) into isolated functions or packages when it's only used once.
**- Justification:** Violates the Rule of Three. Abstracting single-use code adds unnecessary indirection and complexity without improving reusability.
**- Files Affected:** `cmd/redial_proxy/main.go`, `internal/errorreport/*`

## IGNORE: Stating the Obvious in Documentation

**- Pattern:** Adding generic package or method docstrings that merely restate the code's structure (e.g., "Package main is the entrypoint", "main is the entry point").
**- Justification:** The Docs agent must prioritize explaining code nuances and correcting drift. Adding boilerplate docstrings that state the obvious is considered noise.
**- Files Affected:** `cmd/redial_proxy/main.go`, `internal/dialer/dialer.go`, `internal/dialer/dialer_test.go`

## IGNORE: Modifying Build Configurations in Docs PRs

**- Pattern:** Modifying runtime build configurations (like pinning linter versions in `mise.toml`) within a "Docs" Pull Request.
**- Justification:** Docs PRs are strictly limited to documentation changes and must never modify executable logic or runtime build configurations. This mixes concerns and violates scope boundaries.
**- Files Affected:** `mise.toml`

## IGNORE: Ignoring Retroactive Violations

**- Pattern:** Making localized fixes (like changing `time.After` to `time.NewTimer`) without fixing existing violations of the same rule in the same file or surrounding context.
**- Justification:** Agents must prioritize fixing existing retroactive violations. A PR that fixes one instance of a bad pattern while leaving adjacent instances untouched is incomplete and adds noise.
**- Files Affected:** Any file with repeated patterns.

## IGNORE: Misconfigured golangci-lint

**- Pattern:** Removing `golangci-lint` from `[tools]` in `mise.toml` or attempting to install it via `go install` in tasks.
**- Justification:** `golangci-lint` must be managed via `mise` tools and pinned to a specific version to ensure stability and proper integration with the development workflow.
**- Files Affected:** `mise.toml`

## IGNORE: Weakening CI Checks

**- Pattern:** Removing critical steps (like `build`) from the `ci` task in `mise.toml`.
**- Justification:** The `ci` task is explicitly defined to run `lint`, `test`, and `build`. Removing any of these steps compromises the verification process.
**- Files Affected:** `mise.toml`
