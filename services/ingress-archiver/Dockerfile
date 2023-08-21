FROM golang:1.18 AS builder
WORKDIR /sensorbucket

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata -o /ingressarchiver services/ingress-archiver/main.go

FROM scratch AS production
COPY --from=builder /ingressarchiver /ingressarchiver
ENTRYPOINT ["/ingressarchiver"]
