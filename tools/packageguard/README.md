# tools/packageguard

`packageguard` enforces package-size limits to keep packages cohesive and prevent "dumping ground" growth.

## What it checks

- Scans first-level packages under `pkg/`, `services/`, and `tools/`.
- Counts only direct non-test `.go` files in each package directory.
- Fails when a package exceeds the configured max file count.
- Supports an allowlist for intentionally large packages (for example generated code).
- Enforces module topology under `pkg/`:
  - nested module files like `pkg/*/*/go.mod` are rejected
  - `.go` files directly under `pkg/` are rejected

## Usage

```bash
go run ./tools/packageguard -root . -max 12 -allow pkg/protocol
```

## Flags

- `-root`: workspace root path (default `.`)
- `-max`: maximum non-test `.go` files per scanned package (default `12`)
- `-allow`: comma-separated package allowlist (default `pkg/protocol`)

## CI behavior

- Exit code `0`: all checked packages are within limits.
- Exit code `1`: one or more packages exceeded the limit, or scan failed.

Use this together with architecture/package-splitting rules to force early refactoring before packages become mixed-responsibility.
