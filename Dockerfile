FROM golang:1.23-alpine AS build-stage

RUN adduser -D -s /bin/sh uptimego

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN chown -R uptimego:uptimego /app

USER uptimego

RUN CGO_ENABLED=0 GOOS=linux go build -o uptime-go ./cmd

FROM alpine:latest AS release-stage

WORKDIR /

COPY --from=build-stage  /app/uptime-go /
COPY --from=build-stage /app/migrations /

EXPOSE 8080

CMD ["/uptime-go"]