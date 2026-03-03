# tools/protogen

Go executable that generates `pkg/protocol` from protocol YAML.

## Usage

- `go run . -spec ../../vendor/pixel-protocol/spec/protocol.yaml -out ../../pkg/protocol -package protocol`

## Guarantees

- deterministic packet struct generation
- per-realm output files + router map generation
- no runtime business logic in generated files
