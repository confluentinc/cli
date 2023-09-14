FROM confluentinc/confluent-cli:latest

# Switch to root user
USER root

# Copy package.json and package-lock.json to the container
COPY package*.json ./

# Update the package lists for upgrades for security purposes
RUN apk update && apk upgrade

RUN apk update --no-cache && \
    apk add --no-cache curl && \
    apk add --no-cache bash && \
    apk add --no-cache bash-completion && \
    apk add nginx 

RUN apk add --no-cache strace

# Add the official Node.js runtime base image as a source
# Install 'nodejs' and 'npm'
RUN apk add --no-cache nodejs npm

RUN mkdir /certs
COPY privatekey.pem /certs/
COPY cert.pem /certs/

# the following 3 lines are required to run npm install x
# Set the working directory inside the container
WORKDIR /app
RUN chmod +x /bin /etc /app
# Copy package.json and package-lock.json to the container
COPY package*.json ./

# Install server dependencies
RUN npm cache verify
RUN npm install -g npm@latest
# RUN npm ci
RUN npm install ws
RUN npm install fs
RUN npm install https

# Copy the WebSocket server code to the container
COPY server.js ./

# Expose the port on which the WebSocket server will run
EXPOSE 8443
EXPOSE 8080
EXPOSE 8081

# Start the WebSocket server when the container starts
CMD ["node", "server.js"]
