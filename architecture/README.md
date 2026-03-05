# Architecture Directory Contract

This directory is planning-only.

- Use this directory for intended design, decisions, trade-offs, and phased plans.
- Do not describe implementation status here unless explicitly marked as planned or target state.

Global terminology for all files in `architecture/`:

- "service" means internal module/bounded context inside the single `pixelsv` binary.
- "NATS subject" means internal contract topic unless a transport adapter explicitly uses an external broker.

Implemented behavior must be documented under `docs/`.
