FROM golang:1.18-alpine
ENV DEVSVC="Make sure to set DEVSVC env"
WORKDIR /workspace
RUN go install github.com/cespare/reflex@latest
CMD ["sh", "-c", "reflex -r '.go$' -s -t 500ms go run services/$DEVSVC/main.go"]
