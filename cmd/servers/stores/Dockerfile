# syntax=docker/dockerfile:1

# auto-injected BuildKit args
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM

# Build args
ARG GO_VERSION=1.25.9

############################
# Builder stage (Go, glibc)
# IMPORTANT: use TARGETPLATFORM for CGO deps (no CGO cross-compile headaches)
############################
FROM --platform=$TARGETPLATFORM golang:${GO_VERSION}-bookworm AS builder
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG GO_VERSION=1.25.9

WORKDIR /src

RUN echo "Builder stage, fetching tool chain dependencies GO_VERSION=${GO_VERSION} for TARGETARCH=${TARGETARCH}, TARGETOS=${TARGETOS}, TARGETPLATFORM=${TARGETPLATFORM}"

# Toolchain
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates file binutils && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
--mount=type=cache,target=/go/pkg/mod \
GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} \
go build \
-trimpath -buildvcs=false \
-tags netgo,osusergo \
-ldflags "-s -w" \
-o /out/stores-server \
./cmd/servers/stores
# RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} go build -o /out/stores-server ./cmd/servers/stores

# Optional diagnostics
RUN echo "uname:" && uname -a
RUN echo "file:" && file /out/stores-server
RUN echo "interp:" && readelf -l /out/stores-server | grep interpreter || true

############################
# Runtime stage (distroless, glibc)
# IMPORTANT: use base (not static) for shared libs
############################
FROM gcr.io/distroless/base-debian12:nonroot AS runtime
WORKDIR /app

# App
COPY --from=builder /out/stores-server /app/stores-server

# App auth policies (if any)
COPY cmd/servers/stores/policies/model.conf /app/policies/model.conf

EXPOSE 62151
USER nonroot:nonroot
ENTRYPOINT ["/app/stores-server"]

############################
# Debug runtime target (optional)
# Adds a shell via busybox so you can exec in.
# Build with: --target runtime-debug
############################
FROM debian:bookworm-slim AS runtime-debug
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates binutils \
&& rm -rf /var/lib/apt/lists/*

# App
COPY --from=builder /out/stores-server /app/stores-server

# App auth policies (if any)
COPY cmd/servers/stores/policies/model.conf /app/policies/model.conf

# Drop into shell (or change to sleep)
# ENTRYPOINT ["/bin/sh"]
ENTRYPOINT ["/bin/sh", "-c", "sleep 365d"]
