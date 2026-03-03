# pkg/protocol

`pkg/protocol` is generated code for the Pixel Protocol spec at `vendor/pixel-protocol/spec/protocol.yaml`.

## What this module provides

- Header constants (for example `HeaderRoomEnterIn`).
- Strongly typed packet structs per message.
- `Encode` methods and `Decode<PacketName>` functions using `pkg/core/codec`.
- Packet routers and packet-name maps (`C2SRouter`, `S2CRouter`, `C2SPacketName`, `S2CPacketName`).

## Source of truth

- Spec: `vendor/pixel-protocol/spec/protocol.yaml`
- Generator: `tools/protogen`
- Generated files: all `.go` files in this module

Do not hand-edit files in this package. Regeneration will overwrite manual changes.

## Regeneration

From repository root:

```bash
make generate
```

Or directly:

```bash
go run ./tools/protogen -spec vendor/pixel-protocol/spec/protocol.yaml -out pkg/protocol -package protocol
```

## Typical usage

```go
reader := codec.NewReader(payload)
packet, err := protocol.DecodeRoomEnterIn(reader)
if err != nil {
    return fmt.Errorf("decode room.enter: %w", err)
}
roomID := packet.FlatId
_ = roomID
```

## Non-goals

- No business rules.
- No room/service state mutations.
- No custom packet shape changes outside the spec and generator pipeline.
