FROM node:14-alpine

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
    apk add --no-cache curl && \
    apk add --no-cache bash && \
    apk add --no-cache bash-completion && \
    apk add nginx 

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

RUN mkdir /certs
COPY privkey.pem /certs/
COPY cert.pem /certs/

# Install server dependencies
RUN npm install

RUN npm install ws
RUN npm install fs
RUN npm install https

# Copy the WebSocket server code to the container
COPY docker/server.js ./

# Expose the port on which the WebSocket server will run
EXPOSE 443

# Start the WebSocket server when the container starts
CMD ["node", "server.js"]

RUN mkdir /etc/bash_completion.d && confluent completion bash > /etc/bash_completion.d/confluent

USER $USER
