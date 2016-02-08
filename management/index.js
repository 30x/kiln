'use strict'
const env = require('envalid');
const Server = require('./lib/server.js');

const port = env.get('PORT', '8080' );
const tmpDir = env.get('TMP_DIR', '/tmp');

const server = new Server(port, tmpDir);

server.listen();
