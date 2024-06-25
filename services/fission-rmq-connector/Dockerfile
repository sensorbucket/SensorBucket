# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -tags timetzdata \
    -o /fission-rmq-connector \
    services/fission-rmq-connector/main.go

# Final stage
FROM scratch AS production

WORKDIR /

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy the binary from the builder stage
COPY --from=builder /fission-rmq-connector /fission-rmq-connector
# Set timezone environment variable
ENV TZ=Etc/UTC

# Run the binary
ENTRYPOINT ["/fission-rmq-connector"]
