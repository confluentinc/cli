const http = require('http');
const WebSocket = require('ws');
const { exec } = require('child_process');

const httpServer = http.createServer();

const wss = new WebSocket.Server({ server: httpServer });

const textDecoder = new TextDecoder();

console.log('Server script starting up...');

wss.on('connection', (ws, req) => {
  console.log('WebSocket connection established.');

  ws.on('message', (message) => {
    textMessage = textDecoder.decode(message)
    // Check if the message is empty or just whitespace
    if (!textMessage.trim()) {
      console.log('Received an empty message, ignoring it.');
      return; // Exit the function early, not processing the message further
    }
    
    console.log('Received message:', textMessage);

    // Execute the received command in the shell
    exec(textMessage, (error, stdout, stderr) => {
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

wss.on('error', (error) => {
  console.error('WebSocket server error:', error.message);
});

// Set CORS headers for WebSocket handshake response
wss.on('headers', (headers, req) => {
  headers.push('Access-Control-Allow-Origin: *');
  headers.push('Access-Control-Allow-Headers: Origin, X-Requested-With, Content-Type, Accept');
  headers.push('Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS');
});

httpServer.listen(8443, () => {
  console.log('HTTP server listening on port 8443. This is the HTTP server being used for ws connection.');
});
