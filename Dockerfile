# syntax=docker/dockerfile:1.7

ARG GO_BUILD_IMAGE=golang:1.24-alpine3.21
ARG ALPINE_RUNTIME_IMAGE=alpine:3.21

FROM ${GO_BUILD_IMAGE} AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w -buildid=" \
      -o /out/poll-api \
      ./cmd/main.go

FROM ${ALPINE_RUNTIME_IMAGE} AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 appgroup \
    && adduser -S -D -H -u 10001 -G appgroup appuser

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /out/poll-api /app/poll-api

ENV APP_ENV=production \
    SERVER_ADDRESS=:8080

EXPOSE 8080
USER appuser

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD nc -z 127.0.0.1 8080 || exit 1

ENTRYPOINT ["/app/poll-api"]
