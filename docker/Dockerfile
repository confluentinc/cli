FROM alpine:latest

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

COPY confluent /bin

RUN chown $USER:$USER /bin
RUN chown $USER:$USER /etc

# This symbolic link exists for backwards compatibility reasons
RUN ln -s /bin/confluent /confluent

# Install Node.js
RUN curl -sL https://rpm.nodesource.com/setup_14.x | bash -
RUN yum install -y nodejs

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json to the container
COPY package*.json ./

# Install server dependencies
RUN npm install

# Copy the WebSocket server code to the container
COPY docker/server.js ./

# Expose the port on which the WebSocket server will run (e.g., 8080)
EXPOSE 8080

# Start the WebSocket server when the container starts
CMD ["node", "server.js", "&"]

RUN mkdir /etc/bash_completion.d && confluent completion bash > /etc/bash_completion.d/confluent

USER $USER
