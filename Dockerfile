FROM node:10.15.3 AS frontend_builder

COPY ./web/frontend /app/web/frontend
WORKDIR /app/web/frontend

ARG NPM_BUILD_MODE="staging"
RUN npm install
RUN npm run build -- --mode "$NPM_BUILD_MODE"

FROM golang:1.11.8

COPY --from=frontend_builder /app/web/frontend/dist /app/web/frontend/dist
COPY ./datastore /app/datastore
COPY ./handlers /app/handlers
COPY ./types /app/types
COPY ./web/*.go /app/web
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app

ARG GO_BUILD_TAGS="dev"
RUN go build --tags "$GO_BUILD_TAGS" -o /app/main web/main.go

ENTRYPOINT /app/main