# Build the agent binary
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/agent/main.go cmd/agent/main.go
COPY api/ api/

# Build
# CGO_ENABLED=0 for static binary
# Build tags: linux (required for netlink)
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -tags linux -a -o agent cmd/agent/main.go

# Use distroless base image with glibc for compatibility
# The agent needs to run with hostNetwork and elevated privileges to manage ip rules
FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source = "https://github.com/mariusbertram/ip-rule-operator" \
      org.opencontainers.image.description = "Agent for ip-rule-operator managing Linux policy routing rules (ip rule)" \
      org.opencontainers.image.licenses = "Apache-2.0"
WORKDIR /
COPY --from=builder /workspace/agent .
USER 65532:65532

ENTRYPOINT ["/agent"]

