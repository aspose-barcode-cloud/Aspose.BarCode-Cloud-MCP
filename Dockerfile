# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mcp-server .

# Runtime stage
FROM alpine:3.21

# Add ca-certificates for HTTPS calls to Aspose Cloud API
RUN apk --no-cache add ca-certificates

COPY --from=builder /mcp-server /mcp-server

ENTRYPOINT ["/mcp-server"]
CMD ["--mount-path=/mnt/data"]
