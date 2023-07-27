const http = require('http');
const WebSocket = require('ws');
const { exec } = require('child_process');

const server = http.createServer();
const wss = new WebSocket.Server({ server });
const textDecoder = new TextDecoder();

wss.on('connection', (ws) => {
  console.log('WebSocket connection established.');

  ws.on('message', (message) => {
    console.log('Received message:', message);

    // Execute the received command in the shell
    exec(textDecoder.decode(message), (error, stdout, stderr) => {
      if (error) {
        console.error('Error executing command:', error.message);
        ws.send('Error executing command: ' + error.message);
      } else {
        console.log('Command output:', stdout);
        ws.send(stdout);
      }
    });
  });

  ws.on('close', () => {
    console.log('WebSocket connection closed.');
  });
});

server.listen(8080, () => {
  console.log('WebSocket server listening on port 8080.');
});
