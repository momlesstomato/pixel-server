SPEC       := vendor/pixel-protocol/spec/protocol.yaml
PROTO_OUT  := pkg/protocol
PROTO_PKG  := protocol

.PHONY: generate build test lint vet clean docker-up docker-down check-package-split

## Regenerate protocol package from spec YAML.
generate:
	go run ./tools/protogen -spec $(SPEC) -out $(PROTO_OUT) -package $(PROTO_PKG)

## Build every module in the workspace.
build: generate
	go build ./...

## Run all unit tests with race detector.
test:
	go test -race -count=1 ./...

## Run integration tests (requires Docker-backed infra).
test-integration:
	go test -race -count=1 -tags integration ./...

## Run end-to-end tests.
test-e2e:
	go test -race -count=1 -tags e2e ./...

## Run golangci-lint across workspace.
lint:
	golangci-lint run ./...
	go run ./tools/packageguard -root . -max 12 -allow pkg/protocol

## Enforce package splitting limits.
check-package-split:
	go run ./tools/packageguard -root . -max 12 -allow pkg/protocol

## Go vet across workspace.
vet:
	go vet ./...

## Remove build artefacts.
clean:
	rm -f gateway auth game social navigator catalog moderation protogen
	rm -f services/gateway/gateway services/auth/auth services/game/game services/social/social services/navigator/navigator services/catalog/catalog services/moderation/moderation tools/protogen/protogen

## Start local infra with Docker Compose.
docker-up:
	docker compose -f docker/compose.yml up -d

## Stop local infra.
docker-down:
	docker compose -f docker/compose.yml down
