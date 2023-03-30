FROM amd64/alpine:latest

ENV USER=confluent
ENV UID=2358
ENV GID=2358

RUN addgroup \
    --gid $GID \
    --system \
    $USER

RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "$USER" \
    --uid "$UID" \
    "$USER"

RUN apk update --no-cache && \
    apk add --no-cache jq && \
    apk add --no-cache curl && \
    apk add --no-cache bash && \
    apk add --no-cache bash-completion

RUN curl -L --http1.1 https://cnfl.io/cli | sh -s -- -b /bin

RUN chmod +x /bin/confluent
RUN chown $USER:$USER /bin/confluent
RUN chown $USER:$USER /etc

# This symbolic link exists for backwards compatibility reasons
RUN ln -s /bin/confluent /confluent

RUN mkdir /etc/bash_completion.d && confluent completion bash > /etc/bash_completion.d/confluent

USER $USER
