'use strict'

const port = 10001

module.exports = {
  tmpDir: '/tmp',
  port:  port,
  dockerUrl: 'localhost:5000',
  maxFileSize: 104857600,
  url: 'http://localhost:' + port,
  defaultImage: 'node:4.3.0-onbuild'
}


