'use strict';

const app = require('connect')()
const fs = require('fs')
const http = require('http')
const path = require('path')
const swaggerTools = require('swagger-tools')
const YAML = require('js-yaml')

const serverPort = process.env.PORT || 3000
const swaggerDoc = YAML.safeLoad(fs.readFileSync('./config/swagger.yaml', 'utf-8'))

// Initialize the Swagger middleware
swaggerTools.initializeMiddleware(swaggerDoc, (middleware) => {
  // Interpret Swagger resources and attach metadata to request - must be first in swagger-tools middleware chain
  app.use(middleware.swaggerMetadata())

  // Validate Swagger requests
  app.use(middleware.swaggerValidator())

  // Route validated requests to appropriate controller
  app.use(middleware.swaggerRouter({
    controllers: path.join(__dirname, 'controllers')
  }))

  // Serve the Swagger documents and Swagger UI
  app.use(middleware.swaggerUi())

  // Start the server
  http.createServer(app).listen(serverPort, function () {
    console.log('Your server is listening on port %d (http://localhost:%d)', serverPort, serverPort)
  })
})
