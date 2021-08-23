FROM alpine:latest

RUN apk update --no-cache && \
    apk add --no-cache curl && \
    apk add --no-cache bash && \
    apk add --no-cache bash-completion

RUN curl -L --http1.1 https://cnfl.io/cli | sh -s -- -b /bin

RUN chmod +x /bin/cflt

RUN mkdir /etc/bash_completion.d && cflt completion bash > /etc/bash_completion.d/cflt
