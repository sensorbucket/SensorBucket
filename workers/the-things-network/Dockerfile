FROM golang:1.18-alpine AS builder
ARG APP_TYPE="worker"
ARG APP_NAME="the-things-network"
WORKDIR /sensorbucket

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata -o /app ${APP_TYPE}s/${APP_NAME}/main.go

FROM scratch AS server
COPY --from=builder /app /app
ENTRYPOINT [ "/app" ]
