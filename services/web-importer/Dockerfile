FROM oven/bun:latest AS dev
WORKDIR /workspace
EXPOSE 5173

COPY package.json ./
COPY bun.lockb ./

RUN bun install
ENTRYPOINT ["bun", "run", "dev"]


FROM dev AS build
WORKDIR /workspace
COPY . .
RUN bun run build

FROM caddy:alpine AS production
COPY --from=build /workspace/build/ /usr/share/caddy/