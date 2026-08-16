FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/linkbot .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /bin/linkbot /usr/local/bin/linkbot

ENV STORAGE_PATH=/app/data/sqlite/storage.db

VOLUME ["/app/data"]

ENTRYPOINT ["linkbot"]
