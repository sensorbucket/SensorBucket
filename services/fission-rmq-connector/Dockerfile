FROM golang:1.18-alpine AS dev
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

FROM dev AS build
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/fission-rmq-connector /workspace/services/fission-rmq-connector/

FROM scratch AS production
COPY --from=build /app/fission-rmq-connector /fission-rmq-connector
ENTRYPOINT ["/fission-rmq-connector"]
