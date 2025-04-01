FROM oven/bun:latest AS dev
WORKDIR /workspace
EXPOSE 5173

COPY package.json ./
COPY bun.lockb ./

RUN bun install
ENTRYPOINT ["bun", "run", "dev", ""]


