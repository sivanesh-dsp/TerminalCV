# syntax=docker/dockerfile:1

# ---- build stage -------------------------------------------------------------
# Build context is the REPOSITORY ROOT (so we can copy the shared résumé).
# Pin the build stage to the BUILD platform and cross-compile to the TARGET
# platform with Go's GOOS/GOARCH — this avoids slow QEMU emulation of the
# compiler while still producing a correct per-arch binary (CGO is disabled).
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

WORKDIR /src/ssh
# Cache modules first.
COPY ssh/go.mod ssh/go.sum ./
RUN go mod download

# Then the source.
COPY ssh/ ./

ARG VERSION=dev
ARG BUILD_TIME=unknown
# Provided automatically by Buildx for the target platform.
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
      -trimpath \
      -ldflags "-s -w \
        -X github.com/sivanesh/portfolio-ssh/internal/version.Version=${VERSION} \
        -X github.com/sivanesh/portfolio-ssh/internal/version.BuildTime=${BUILD_TIME}" \
      -o /out/portfolio-ssh ./cmd/portfolio-ssh

# ---- runtime stage -----------------------------------------------------------
# No --platform here: Buildx resolves alpine for each TARGET platform, so the
# resulting image is native to arm64 / amd64 respectively.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget su-exec \
 && addgroup -S app && adduser -S -G app app \
 && mkdir -p /data /var/log/portfolio && chown app:app /data /var/log/portfolio

COPY --from=build /out/portfolio-ssh /app/portfolio-ssh
COPY ssh/docker-entrypoint.sh /app/docker-entrypoint.sh
# The single shared source of truth (consumed at runtime).
COPY shared/resume.json /app/shared/resume.json
RUN chmod +x /app/docker-entrypoint.sh

ENV RESUME_PATH=/app/shared/resume.json \
    SSH_ADDR=:2222 \
    HEALTH_ADDR=:8081 \
    HOST_KEY_PATH=/data/host_ed25519 \
    STATE_PATH=/data/state.json

# NOTE: no `USER app` here — the entrypoint starts as root only to fix bind-mount
# ownership, then drops to the `app` user via su-exec before running the server.
VOLUME ["/data"]
EXPOSE 2222 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8081/healthz || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
