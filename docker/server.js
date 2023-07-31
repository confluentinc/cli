const https = require('https');
const fs = require('fs');
const WebSocket = require('ws');
const { exec } = require('child_process');

const privateKeyPath = '/certs/privkey.pem';
const certificatePath = '/certs/cert.pem';
const sslCertDir = process.env.SSL_CERT_DIR || '/certs';

const privateKey = fs.readFileSync(privateKeyPath, 'utf8');
const certificate = fs.readFileSync(certificatePath, 'utf8');

const credentials = { key: privateKey, cert: certificate };

const httpsServer = https.createServer(credentials, (req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello, this is an HTTPS server.');
});

const wss = new WebSocket.Server({ httpsServer });

const textDecoder = new TextDecoder();

wss.on('connection', (ws, req) => {
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

wss.on('error', (error) => {
  console.error('WebSocket server error:', error.message);
});

// Set CORS headers for WebSocket handshake response
wss.on('headers', (headers, req) => {
  headers.push('Access-Control-Allow-Origin: *');
  headers.push('Access-Control-Allow-Headers: Origin, X-Requested-With, Content-Type, Accept');
  headers.push('Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS');
});

httpsServer.listen(443, () => {
  console.log('HTTPS server listening on port 443.');
});