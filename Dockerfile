FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26.5 AS build
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
ENV CGO_ENABLED=0
RUN go build -o /game-server ./cmd/server

FROM alpine:3.22
COPY --from=build /game-server /game-server
COPY --from=web /src/web/dist /web
ENV GAME_ADDR=0.0.0.0:8080
ENV WEB_ROOT=/web
EXPOSE 8080
USER nobody
ENTRYPOINT ["/game-server"]
