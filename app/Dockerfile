# Stage 1: Builder
# Used a pinned version of Go with Alpine for a smaller build context
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Optimization: Copy dependencies manifest first to leverage Docker cache
# This layer will only rebuild if go.mod or go.sum changes
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
# CGO_ENABLED=0: Ensures a statically linked binary (no dependency on system libc)
# -ldflags="-s -w": Strips debug information (reduces size and makes reverse engineering harder)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server main.go

# Stage 2: Runner (Distroless) 
# Use 'static' distroless image (minimal attack surface, no shell, no package manager)
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Security: Run as non-root user (uid: 65532 is the default nonroot user in distroless)
USER 65532:65532

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# Documentation: Expose the port the app listens on
EXPOSE 8080

CMD ["./server"]