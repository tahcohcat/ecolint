# Dockerfile.alpine - Alternative with Alpine base for more features
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

WORKDIR /src

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-w -s' -o ecolint cmd/ecolint/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and basic shell utilities
RUN apk --no-cache add ca-certificates bash git

# Create non-root user
RUN adduser -D -s /bin/bash ecolint

# Copy binary
COPY --from=builder /src/ecolint /usr/local/bin/ecolint

# Set ownership and permissions
RUN chown ecolint:ecolint /usr/local/bin/ecolint && \
    chmod +x /usr/local/bin/ecolint

# Switch to non-root user
USER ecolint

# Set working directory
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/ecolint"]

---