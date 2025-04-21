# Use Go base image for building
FROM golang:1.22.2 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the full source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o llm-orchestrator ./main.go

# Use a lightweight final image
FROM alpine:latest

WORKDIR /root/

# Copy binary and any required files
COPY --from=builder /app/llm-orchestrator .
COPY data ./data
COPY internal/promptbuilder ./internal/promptbuilder

# Expose the port and run the binary
EXPOSE 8080
CMD ["./llm-orchestrator"]
