FROM alpine:latest

WORKDIR /app

COPY ./bin/discord-bot .

RUN chmod +x ./discord-bot

CMD ["./discord-bot"]