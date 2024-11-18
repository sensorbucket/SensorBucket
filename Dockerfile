ARG APP_NAME
ARG APP_TYPE
FROM golang:1.22-alpine AS dev
ENV APP_NAME=${APP_NAME}
ENV APP_TYPE=${APP_TYPE}
WORKDIR /workspace

RUN go install github.com/cespare/reflex@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["sh", "-c", "reflex -r '.(go|html)$' -s go run ./${APP_TYPE}s/${APP_NAME}"]

FROM dev AS build
ARG APP_NAME
ARG APP_TYPE
WORKDIR /workspace

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/${APP_NAME} ${APP_TYPE}s/${APP_NAME}/main.go

FROM scratch AS production
ARG APP_NAME
ARG APP_TYPE
COPY --from=build /app/${APP_NAME} /app/${APP_NAME}
ENTRYPOINT /app/${APP_NAME}

