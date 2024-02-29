FROM golang:1.21 AS builder
WORKDIR /sensorbucket

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata -o /core services/core/main.go

FROM scratch AS production
COPY --from=builder /core /core
ENTRYPOINT ["/core"]
