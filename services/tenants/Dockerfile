# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

RUN go install github.com/valyala/quicktemplate/qtc@latest
RUN go install github.com/evanw/esbuild/cmd/esbuild@latest
RUN apk add --no-cache git ca-certificates tzdata make curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.3/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/bin/tailwind \
    && ln -s /usr/bin/tailwind /usr/bin/tailwindcss

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

RUN make -C services/tenants build

# Build the application
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -tags timetzdata \
    -o /tenants \
    services/tenants/main.go

# Final stage
FROM scratch AS production

WORKDIR /

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy the binary from the builder stage
COPY --from=builder /tenants /tenants
# Set timezone environment variable
ENV TZ=Etc/UTC

# Run the binary
ENTRYPOINT ["/tenants"]

