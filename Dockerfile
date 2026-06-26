# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.25-alpine AS builder

# Build-time metadata, injected into the version package.
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a static binary.
COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath \
    -ldflags "-s -w \
      -X github.com/Skypieee6/redintel-sentinel/internal/version.Version=${VERSION} \
      -X github.com/Skypieee6/redintel-sentinel/internal/version.Commit=${COMMIT} \
      -X github.com/Skypieee6/redintel-sentinel/internal/version.BuildTime=${BUILD_TIME}" \
    -o /out/redintel ./cmd/server

############################
# Runtime stage
############################
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 redintel

WORKDIR /app

COPY --from=builder /out/redintel /usr/local/bin/redintel
COPY --from=builder /src/configs ./configs
COPY --from=builder /src/migrations ./migrations

USER redintel

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["redintel"]
CMD ["serve"]
