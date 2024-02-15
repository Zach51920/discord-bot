FROM alpine:latest

WORKDIR /app

COPY ./bin/youpirate .

CMD ["./youpirate"]