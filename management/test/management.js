'use strict'
const should = require('should')
const restify = require('restify')
const fs = require('fs')
const http = require('http')
const Server = require('../lib/server.js')
const Docker = require('../lib/docker.js')


const port = 10001
const tmpDir = '/tmp'


const dockerUrl =  'localhost:5000'

const docker = new Docker(dockerUrl)

const maxFileSize = 104857600

const server = new Server(port, tmpDir, maxFileSize, docker)

const url = 'http://localhost:' + port

const jsonClient = restify.createJsonClient({
  url: url,
  version: '~1.0'
})


describe('management', function () {
  /**
   * Start and stop the server before and after the test suites
   */
  before(function () {
    server.listen()
  })

  after(function () {
    server.close()
  })

  describe('#heartbeat()', function () {
    it('should get without error', function (done) {
      jsonClient.get('/v1/heartbeat', function (err, req, res, data) {
        if (err) {
          throw new Error(err)
        }

        should(res).not.null()
        should(res.statusCode).not.null()
        should(res.statusCode).not.undefined()
        res.statusCode.should.equal(200)


        should(data).not.null()
        should(data.status).not.undefined()
        data.status.should.equal('ok')


        done()

      })
    })
  })

  describe('#upload()', function () {
    it('happy path with zip', function (done) {

      const client = new Client('localhost', port)

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/echo-test.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err)
        }
        should(response).not.null()
        should(response.statusCode).not.null()
        should(response.statusCode).not.undefined()
        response.statusCode.should.equal(200)


        const json = JSON.parse(bodyBuffer)

        should(json).not.null()
        should(json.endpoint).not.undefined()
        json.endpoint.should.equal('http://endpointyouhit:8080')

        done()

      })
    })

    //test a valid zip with no package.json file
    it('valid zip no package.json ', function (done) {

      const client = new Client('localhost', port)

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/text-file.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err)
        }

        should(response).not.null()
        should(response.statusCode).not.null()
        should(response.statusCode).not.undefined()
        response.statusCode.should.equal(400)

        should(bodyBuffer).not.null()

        const json = JSON.parse(bodyBuffer)

        should(json.message).not.null()

        json.message.should.equal("Unable to validate node application. package.json could not be read.  Ensure it is in your upload.")

        done()


      })
    })

    //test incorrect zip file encoding
    it('not a valid zip ', function (done) {

      const client = new Client('localhost', port)

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/not-a-zip.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err)
        }

        should(response).not.null()
        should(response.statusCode).not.null()
        should(response.statusCode).not.undefined()
        response.statusCode.should.equal(400)


        should(bodyBuffer).not.null()

        const json = JSON.parse(bodyBuffer)

        should(json.message).not.null()

        json.message.should.equal("Unable to extract zip file.  Ensure you have a valid zip file.")

        done()

      })
    })

    //test there's a run command in the package json
    it('no run in package.json ', function (done) {

      const client = new Client('localhost', port)

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/echo-test-no-run.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err)
        }

        should(response).not.null()
        should(response.statusCode).not.null()
        should(response.statusCode).not.undefined()
        response.statusCode.should.equal(400)


        should(bodyBuffer).not.null()

        const json = JSON.parse(bodyBuffer)

        should(json.message).not.null()

        json.message.should.equal("Unable to validate node application. scripts.start is required in package.json.")

        done()

      })
    })




    //TODO test file too large
    //TODO Test missing header


  })
})


/**
 *
 * @constructor
 */
const Client = function (host, port) {
  this.host = host
  this.port = port
}

/**
 * Post the zip file.  Callback is of the type
 *
 * function(err, response, bodyBuffer)
 */
Client.prototype.putZipFile = function (orgName, envName, appName, revision, zipFilePath, cb) {

  fs.existsSync(zipFilePath).should.equal(true)

  const readStream = fs.createReadStream(zipFilePath)

  const headers = {}
  headers['x-apigee-script-container-rev'] = revision

  var options = {
    host: 'localhost'
    , port: port
    , path: '/v1/buildnodejs/' + orgName + '/' + envName + '/' + appName
    , method: 'PUT'
    , headers: headers
  }

  var req = http.request(options, function (res) {
    //this feels a bit backwards, but these are evaluated AFTER the read stream has closed

    var buffer = ''

    //pipe body to a buffer
    res.on('data', function (data) {
      buffer += data
    })

    res.on('end', function () {

      cb(null, res, buffer)

    })

  })

  req.on('error', function (err) {
    if (err) {
      cb(err, null, null)
    }

  })

  //pipe the readstream into the request
  readStream.pipe(req)


  /**
   * Close the request on the close of the read stream
   */
  readStream.on('close', function () {
    req.end()
    console.log('I finished.')
  })


  //note that if we end up with larger files, we may want to support the continue, much as S3 does
  //https://nodejs.org/api/http.html#http_event_continue


}

