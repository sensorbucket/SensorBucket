FROM node:18-alpine AS dev
WORKDIR /app
COPY . .
RUN yarn set version stable && \
    yarn install && \
    yarn build

FROM node:18-alpine AS prod
ENV PORT=3000
ENV HOST=0.0.0.0
WORKDIR /app
COPY --from=dev /app/.yarn /app/.yarn
COPY --from=dev /app/package.json /app/package.json 
COPY --from=dev /app/build /app/build 
CMD yarn node build
