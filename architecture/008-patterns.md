# Architectural Patterns

## Mandatory Patterns

- Hexagonal Architecture (Ports and Adapters)
- DDD (bounded contexts, aggregates, ubiquitous language)
- TDD for domain and reusable core logic
- ECS for realtime simulation concerns

## Hexagonal Rules

- Domain packages define ports they consume.
- Adapters implement ports and remain outside domain.
- Entrypoints (HTTP, WebSocket, CLI, jobs) call application services, not domain internals directly.

## DDD Rules

- Model invariants inside aggregate roots.
- Keep bounded contexts explicit (`auth`, `game`, `social`, `catalog`, `navigator`, `moderation`).
- Cross-context communication uses contracts and IDs, not direct struct sharing.

## TDD Rules

- Write tests before or alongside domain behavior.
- Prioritize table-driven tests for deterministic rules.
- Adapter concerns use integration tests.
- End-to-end scenarios validate full binary flows.

## ECS Rules

- ECS is required for room/world simulation and entity updates.
- Room workers own ECS state and enforce single-writer mutation.
- External events become commands/envelopes, never direct ECS writes from outside the owner loop.

## Reusability Rules

- Prefer `pkg/` reusable packages for transport-agnostic logic.
- Avoid framework dependencies in reusable packages.
- Keep APIs small and stable.

## Anti-Patterns

- Domain logic in handlers/controllers.
- Shared mutable state across module boundaries.
- Concrete infrastructure dependencies imported into domain code.
- Copy-pasted packet/business logic between modules.
