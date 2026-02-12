# ── Build stage ──────────────────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

RUN apk add --no-cache upx ca-certificates

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -trimpath -o /pluma ./cmd/pluma

# Compress binary with UPX for maximum size reduction
RUN upx --best --lzma /pluma

# ── Final stage (scratch = 0 bytes base) ────────────────────
FROM scratch

# TLS certificates for HTTPS calls to Telegram API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary
COPY --from=builder /pluma /pluma
COPY config.yaml /config.yaml


# Default config path
ENV CONFIG_PATH=/config.yaml

EXPOSE 8080

ENTRYPOINT ["/pluma"]
