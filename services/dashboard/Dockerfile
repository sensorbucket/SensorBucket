# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS dev

# Set working directory
WORKDIR /workspace

RUN go install github.com/valyala/quicktemplate/qtc@latest
RUN go install github.com/evanw/esbuild/cmd/esbuild@latest
RUN go install github.com/cespare/reflex@latest
RUN apk add --no-cache git ca-certificates tzdata make curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.3/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/bin/tailwind \
    && ln -s /usr/bin/tailwind /usr/bin/tailwindcss

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

CMD make -C services/dashboard watch

FROM dev AS builder
WORKDIR /app

# Copy the source code
COPY . .

RUN make -C services/dashboard build

# Build the application
ARG TARGETOS TARGETARCH GITVERSION
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    GITVERSION=${GITVERSION} GOVERSION=$(go version) BUILDDATE=$(date -Is) \
    go build \
    -ldflags="-w -s -extldflags \"-static\" \
    -X 'sensorbucket.nl/sensorbucket/internal/buildinfo.GitVersion=${GITVERSION}' \
    -X 'sensorbucket.nl/sensorbucket/internal/buildinfo.BuildTime=$(date -Is)' \
    -X 'sensorbucket.nl/sensorbucket/internal/buildinfo.Architecture=${TARGETARCH}' \
    -X 'sensorbucket.nl/sensorbucket/internal/buildinfo.GoVersion=$(go version)'" \
    -tags timetzdata \
    -o /tenants \
    services/dashboard/main.go

# Final stage
FROM scratch AS production

WORKDIR /

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy the binary from the builder stage
COPY --from=builder /tenants /app
# Set timezone environment variable
ENV TZ=Etc/UTC

# Run the binary
ENTRYPOINT ["/app"]

