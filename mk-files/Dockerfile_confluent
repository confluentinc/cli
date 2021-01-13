FROM alpine:latest

RUN apk update --no-cache && \
    apk add --no-cache curl

RUN curl -L --http1.1 https://cnfl.io/cli | sh -s -- -b .

RUN chmod +x confluent

