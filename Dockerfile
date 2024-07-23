FROM golang:1.21-alpine AS build

WORKDIR /app

COPY . .

RUN go build -o discord-bot .

FROM alpine:latest

WORKDIR /app

ARG BOT_TOKEN
ARG GOOGLE_API_KEY

ENV BOT_TOKEN=${BOT_TOKEN}
ENV GOOGLE_API_KEY=${GOOGLE_API_KEY}

COPY --from=build /app/discord-bot .
COPY --from=build /app/config.yaml /app/config.yaml
COPY --from=build /app/migrations /app/migrations

CMD [ "./discord-bot" ]
