# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Pre-copy and cache Go modules dependency layers
COPY go.mod go.sum .env ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/main.go

# Stage 2: Create a minimal deployment image
FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /main /main
COPY --from=builder /app/.env ./

ENTRYPOINT ["/main"]
