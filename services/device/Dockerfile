FROM golang:1.18 AS builder
WORKDIR /sensorbucket

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata -o /device services/device/main.go

FROM scratch AS production
COPY --from=builder /device /device
ENTRYPOINT ["/device"]
