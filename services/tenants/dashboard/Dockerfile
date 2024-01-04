FROM golang:1.21-alpine AS dev
WORKDIR /workspace

RUN go install github.com/cespare/reflex@latest
RUN go install github.com/valyala/quicktemplate/qtc@latest
RUN go install github.com/evanw/esbuild/cmd/esbuild@latest
RUN apk add make curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.3/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/bin/tailwind \
    && ln -s /usr/bin/tailwind /usr/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download

CMD ["make", "watch-dashboard"]

FROM dev AS build
WORKDIR /workspace

COPY . .
RUN GOOS=linux go build -o /app/tenants/dashboard ./services/tenants/dashboard

FROM alpine AS production
COPY --from=build /app/tenants/dashboard /app/tenants/dashboard
ENTRYPOINT ["/app/tenants/dashboard"]