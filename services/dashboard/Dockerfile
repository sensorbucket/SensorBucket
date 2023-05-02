FROM node:18-alpine as builder
WORKDIR /app

COPY . .
RUN yarn install && yarn build && yarn prod-install build

FROM node:18-alpine as production
WORKDIR /app
COPY --from=builder /app/build /app

EXPOSE 3000
CMD ["yarn", "node", "."]

