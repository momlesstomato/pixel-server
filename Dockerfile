# ── Builder ──────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY pkg/        pkg/
COPY tools/      tools/
COPY services/   services/

RUN go build -o /out/gateway    ./services/gateway
RUN go build -o /out/auth       ./services/auth
RUN go build -o /out/game       ./services/game
RUN go build -o /out/social     ./services/social
RUN go build -o /out/navigator  ./services/navigator
RUN go build -o /out/catalog    ./services/catalog
RUN go build -o /out/moderation ./services/moderation

# ── Runtime images ───────────────────────────────────────────────────────────
FROM alpine:3.20 AS gateway
COPY --from=builder /out/gateway /app/gateway
ENTRYPOINT ["/app/gateway"]

FROM alpine:3.20 AS auth-svc
COPY --from=builder /out/auth /app/auth
ENTRYPOINT ["/app/auth"]

FROM alpine:3.20 AS game-svc
COPY --from=builder /out/game /app/game
ENTRYPOINT ["/app/game"]

FROM alpine:3.20 AS social-svc
COPY --from=builder /out/social /app/social
ENTRYPOINT ["/app/social"]

FROM alpine:3.20 AS navigator-svc
COPY --from=builder /out/navigator /app/navigator
ENTRYPOINT ["/app/navigator"]

FROM alpine:3.20 AS catalog-svc
COPY --from=builder /out/catalog /app/catalog
ENTRYPOINT ["/app/catalog"]

FROM alpine:3.20 AS moderation-svc
COPY --from=builder /out/moderation /app/moderation
ENTRYPOINT ["/app/moderation"]
