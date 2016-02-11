'use strict'
const env = require('envalid')
const Server = require('./lib/server.js')

const port = env.get('PORT', '8080' )
const tmpDir = env.get('TMP_DIR', '/tmp')

/**
 * 100 megs is our default max
 * 100*1024*1024 =  104857600
 */

const maxFileSize = env.get('MAX_UPLOAD_SIZE','104857600')

const server = new Server(port, tmpDir, maxFileSize)

server.listen()
