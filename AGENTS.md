# Pixel Server Emulator: Engineering Contract

This document defines mandatory project constraints and success criteria.

## 1) Architecture and Design

- DDD and Hexagonal Architecture are mandatory and heavily enforced.
- DDD + Hexagonal compliance is always a success criterion for delivered work.
- Repository structure must include:
  - `core/` for essential platform and application-wide capabilities.
  - `pkg/` for realm-only logic.
- Domain ownership belongs to each realm present in the pixel protocol.
- Code must follow HashiCorp and Linux philosophy:
  - small focused units,
  - composability,
  - explicit interfaces and contracts,
  - predictable operational behavior.

## 2) Documentation Standards

- All Go code must use GoDoc-style documentation.
- Every method signature must be documented.
- Every interface signature must be documented.
- Every struct signature must be documented.
- Every test function signature must be documented.
- Every struct field and interface property must be documented.
- No comments are allowed inside function bodies.

## 3) Runtime and Interfaces

- Code must be performance-oriented, asynchronous-focused, and self-explanatory.
- The stack must use:
  - Fiber for HTTP API.
  - Fiber WebSocket for realtime communication.
  - Cobra for CLI.
- API and CLI capabilities must remain 1:1 in behavior and feature surface.

## 4) Protocol Source of Truth

- The `vendor/` directory contains:
  - the protocol implementation being targeted,
  - four legacy protocol versions.
- Emulator protocol behavior must be derived from these vendor artifacts.

## 5) Data and Persistence

- PostgreSQL and Redis are required.
- PostgreSQL schema must be normalized.
- PostgreSQL access must use an ORM.
- Redis is the mandatory caching/runtime state layer where appropriate.

## 6) Configuration Module Rules

- A `core/config` package is mandatory and must use Viper.
- Configuration must be parsed from both `.env` file and environment variables.
- Configuration must be section-structured using composed structs.
- Base application config must include app-level settings such as bind IP and port.
- Base application config must compose dedicated sections (Redis, PostgreSQL, Users, and others as needed).
- Any struct property with a `default` Go tag uses that default value.
- Any struct property without a `default` tag is mandatory.
- Startup must fail when a mandatory configuration value is missing.

## 7) Logging Module Rules

- A `core/logging` package is mandatory and the entire app must use Zap.
- Logging output must be switchable between `json` and `console`.
- Logging level must be configurable via environment variable.

## 8) Documentation Folder Policy

- A top-level `docs/` folder is mandatory.
- Documentation style must follow a wiki approach, inspired by:
  - https://github.com/google/guice/wiki/LinkedBindings
  - the Guice wiki structure and clarity principles.
- Documentation in the wiki must only be written when explicitly ordered.

## 9) Testing and Size Constraints

- Every delivered piece must include corresponding automated test coverage.
- Coverage target is 100% or as close as possible, including edge and extreme cases.
- Source files must not exceed 150 lines of code (documentation excluded).
- Packages must not exceed 6 files.
- If a test exceeds 150 lines of code, that package must use an internal `tests/` folder.
- Test files inside `tests/` must use descriptive names and never generic splits like `part1` or `part2`.
