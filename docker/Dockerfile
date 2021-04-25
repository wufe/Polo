FROM node:14 AS frontend-builder
WORKDIR /app
COPY . /app
RUN npm install && \
    npm run clean && \
    npm run build:client:prod

FROM golang:1.16 AS backend-builder
WORKDIR /app
COPY . /app
COPY --from=frontend-builder /app/public /app/public
RUN go get github.com/rakyll/statik && \
    go mod download && \
    statik -p=services -dest=./pkg -src=./public && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o release/polo-linux-amd64 cmd/server/main.go

FROM docker:20-git
LABEL maintainer="simone.bembi@gmail.com"

COPY --from=backend-builder /app/release/polo-linux-amd64 /app/polo
WORKDIR /app
RUN chmod +x polo

ENTRYPOINT ./polo