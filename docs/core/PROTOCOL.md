# PROTOCOL

## Overview

Protocol implementation is split into:

- `pkg/codec` for binary primitives and frame parsing.
- `pkg/protocol` for generated packet structs and c2s decoders.
- `tools/protogen` for YAML-to-Go generation from `vendor/pixel-protocol/spec/protocol.yaml`.

## Framing

Implemented framing matches the protocol spec:

- `uint32` big-endian length prefix.
- `uint16` big-endian message header.
- payload bytes in declared field order.
- multiple packets can be concatenated in one websocket payload.

`pkg/codec` APIs:

- `EncodeFrame(header, payload)`
- `SplitFrames(raw)`
- `Reader` / `Writer` primitives (`int32`, `uint16`, `uint32`, `string`, `bytes`, `bool`)

## Generated Contracts

Current generated scope:

- realm: `handshake-security`
- direction: `c2s`
- file: `pkg/protocol/handshake_security_c2s_gen.go`

Decoder registry entrypoint:

- `protocol.DecodeC2S(header, payload)`

## Generation Command

```bash
go generate ./pkg/protocol
```

Directive source:

- `pkg/protocol/doc.go`
