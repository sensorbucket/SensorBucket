FROM golang:1.19 AS builder
WORKDIR /sensorbucket

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata -o /user-workers services/fission-user-workers/main.go

FROM scratch AS production
COPY --from=builder /user-workers /user-workers
ENTRYPOINT ["/user-workers"]
