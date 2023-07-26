FROM node:14

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

RUN apt-get update && apt-get install -y curl && \
    apt-get install -y jq && \
    apt-get install -y bash && \
    apt-get install -y bash-completion

RUN curl -L --http1.1 https://cnfl.io/cli | sh -s -- -b /bin

RUN chmod +x /bin/confluent
RUN chown $USER:$USER /bin/confluent

RUN chown $USER:$USER /bin
RUN chown $USER:$USER /etc

# This symbolic link exists for backwards compatibility reasons
RUN ln -s /bin/confluent /confluent

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
