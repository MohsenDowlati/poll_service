FROM golang:1.19-alpine AS builder

# Install build deps
RUN apk add --no-cache git ca-certificates build-base

WORKDIR /app

# cache modules
COPY go.mod go.sum ./
RUN go mod download

# copy source
COPY . .

# Build statically to keep runtime image small
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /app/bin/main ./cmd/main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# copy binary from builder
COPY --from=builder /app/bin/main /app/main

EXPOSE 8080

ENTRYPOINT ["/app/main"]