'use strict'
const util = require('util')
const Server = require('./lib/server.js')
const Docker = require('./lib/docker.js')
const env = require('./lib/env.js')


console.log('Environment is %s', util.inspect(process.env))


const port = env.get('PORT', '3000')
const tmpDir = env.get('TMP_DIR', '/tmp')

/**
 * 100 megs is our default max
 * 100*1024*1024 =  104857600
 */


const maxFileSize = env.get('MAX_UPLOAD_SIZE', '104857600')

const dockerUrl = env.get('SHIPYARD_REPO', 'localhost:5000')

const defaultImage = env.get('SHIPYARD_DEFAULT_IMAGE', 'node:4.3.0-slim')

const docker = new Docker(dockerUrl)

console.log('initializing docker images')

docker.initialize(defaultImage, function (err) {

  if(err){
    throw err
  }

  console.log('starting server')

  const server = new Server(port, tmpDir, maxFileSize, docker)

  server.listen()

})
