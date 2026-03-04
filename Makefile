SPEC       := vendor/pixel-protocol/spec/protocol.yaml
PROTO_OUT  := pkg/protocol
PROTO_PKG  := protocol
MODULE_DIRS := $(shell find . -name go.mod -not -path './vendor/*' -exec dirname {} \; | sort)

.PHONY: generate build test lint vet clean docker-up docker-down check-package-split

## Regenerate protocol package from spec YAML.
generate:
	go run ./tools/protogen -spec $(SPEC) -out $(PROTO_OUT) -package $(PROTO_PKG)

## Build every module in the workspace.
build: generate
	@set -e; for d in $(MODULE_DIRS); do \
		name=$$(basename "$$d"); \
		echo "==> build $$d"; \
		(cd "$$d" && go build ./... && rm -f "$$name"); \
	done

## Run all unit tests with race detector.
test:
	@set -e; for d in $(MODULE_DIRS); do \
		echo "==> test $$d"; \
		(cd $$d && go test -race -count=1 ./...); \
	done

## Run integration tests (requires Docker-backed infra).
test-integration:
	@set -e; for d in $(MODULE_DIRS); do \
		echo "==> integration $$d"; \
		(cd $$d && go test -race -count=1 -tags integration ./...); \
	done

## Run end-to-end tests.
test-e2e:
	@set -e; for d in $(MODULE_DIRS); do \
		echo "==> e2e $$d"; \
		(cd $$d && go test -race -count=1 -tags e2e ./...); \
	done

## Run golangci-lint across workspace.
lint:
	@set -e; for d in $(MODULE_DIRS); do \
		echo "==> lint $$d"; \
		(cd $$d && golangci-lint run ./...); \
	done
	go run ./tools/packageguard -root . -max 12 -allow pkg/protocol

## Enforce package splitting limits.
check-package-split:
	go run ./tools/packageguard -root . -max 12 -allow pkg/protocol

## Go vet across workspace.
vet:
	@set -e; for d in $(MODULE_DIRS); do \
		echo "==> vet $$d"; \
		(cd $$d && go vet ./...); \
	done

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
