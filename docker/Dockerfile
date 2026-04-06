# ─────────────────────────────────────────────────────────────────────────────
# Stage 1 — deps
#   Download and cache Go modules separately so that a code-only change
#   doesn't re-download the entire dependency tree.
# ─────────────────────────────────────────────────────────────────────────────
FROM golang:latest AS deps

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify


# ─────────────────────────────────────────────────────────────────────────────
# Stage 2 — builder
#   Compile a fully-static binary. CGO is disabled so the binary runs on
#   the distroless/scratch image in the final stage without any C runtime.
# ─────────────────────────────────────────────────────────────────────────────
FROM deps AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w" \
      -o /app/bin/server \
      ./cmd/api


# ─────────────────────────────────────────────────────────────────────────────
# Stage 3 — final
#   Minimal runtime image — no shell, no package manager, no attack surface.
#   Only the compiled binary and the timezone / CA-cert data it needs.
# ─────────────────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS final

# Run as non-root user (distroless provides uid 65532 "nonroot")
USER nonroot:nonroot

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder --chown=nonroot:nonroot /app/bin/server ./server

EXPOSE 8080

ENTRYPOINT ["./server"]
