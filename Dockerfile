# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries for all three examples
RUN go build -o /app/bin/basic ./examples/basic/main.go
RUN go build -o /app/bin/manager ./examples/manager/main.go
RUN go build -o /app/bin/batch ./examples/batch/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/bin/ ./bin/

# Copy source for reference
COPY --from=builder /app/ ./app/
