FROM alpine:latest

RUN apk update --no-cache && \
    apk add --no-cache curl && \
    apk add --no-cache bash && \
    apk add --no-cache bash-completion

RUN curl -L --http1.1 https://cnfl.io/cli | sh -s -- -b /bin

RUN chmod +x /bin/confluent

# This symbolic link exists for backwards compatibility reasons
RUN ln -s /bin/confluent /confluent

RUN mkdir /etc/bash_completion.d && confluent completion bash > /etc/bash_completion.d/confluent
