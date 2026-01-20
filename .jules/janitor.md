# Janitor's Journal

This journal is for recording critical learnings about refactoring in this codebase.

## 2026-01-20 - Fix blocking dial in redial proxy

**Issue:** The `redial` function used `net.Dial` which blocks until system timeout, ignoring the provided context. It also used a busy-wait-like `select` structure.
**Root Cause:** `net.Dial` does not accept a context. The loop structure was checking context only between dial attempts.
**Solution:** Replaced `net.Dial` with `net.Dialer{}.DialContext`. Simplified loop to respect context cancellation immediately and during retry sleep.
**Pattern:** Always use `DialContext` (or context-aware functions) in long-running operations or network calls to ensure responsiveness to shutdown/timeout signals.
