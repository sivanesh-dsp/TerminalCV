# syntax=docker/dockerfile:1

# ---- build stage -------------------------------------------------------------
# Build context is the REPOSITORY ROOT (so we can copy the shared résumé).
FROM golang:1.26-alpine AS build

WORKDIR /src/ssh
# Cache modules first.
COPY ssh/go.mod ssh/go.sum ./
RUN go mod download

# Then the source.
COPY ssh/ ./

ARG VERSION=dev
ARG BUILD_TIME=unknown
ENV CGO_ENABLED=0
RUN go build \
      -trimpath \
      -ldflags "-s -w \
        -X github.com/sivanesh/portfolio-ssh/internal/version.Version=${VERSION} \
        -X github.com/sivanesh/portfolio-ssh/internal/version.BuildTime=${BUILD_TIME}" \
      -o /out/portfolio-ssh ./cmd/portfolio-ssh

# ---- runtime stage -----------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget \
 && addgroup -S app && adduser -S -G app app \
 && mkdir -p /data && chown app:app /data

COPY --from=build /out/portfolio-ssh /app/portfolio-ssh
# The single shared source of truth (consumed at runtime).
COPY shared/resume.json /app/shared/resume.json

ENV RESUME_PATH=/app/shared/resume.json \
    SSH_ADDR=:2222 \
    HEALTH_ADDR=:8081 \
    HOST_KEY_PATH=/data/host_ed25519 \
    STATE_PATH=/data/state.json

USER app
VOLUME ["/data"]
EXPOSE 2222 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8081/healthz || exit 1

ENTRYPOINT ["/app/portfolio-ssh"]
