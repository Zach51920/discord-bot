FROM golang:1.21-alpine AS build

WORKDIR /app

COPY . .

RUN go build -o discord-bot .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/discord-bot .
COPY --from=build /app/migrations /app/migrations

CMD [ "./discord-bot" ]
