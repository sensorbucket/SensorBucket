FROM golang:1.18-alpine
ENV APP_NAME="Make sure to set APP_NAME env"
ENV APP_TYPE="service"
WORKDIR /workspace
RUN go install github.com/cespare/reflex@latest
CMD ["sh", "-c", "reflex -r '.go$' -s -t 500ms go run ${APP_TYPE}s/${APP_NAME}/main.go"]
