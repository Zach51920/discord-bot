FROM golang:1.21-alpine AS Build

WORKDIR /app

COPY . .

RUN go build -o discord-bot .

FROM alpine:latest

WORKDIR /app

COPY --from=Build /app/discord-bot .

CMD [ "./discord-bot" ]
