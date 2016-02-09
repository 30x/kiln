'use strict'
const should = require('should');
const restify = require('restify');
const fs = require('fs');
const http = require('http');
const Server = require('../lib/server.js');


const port = 8080;
const tmpDir = '/tmp';

const server = new Server(port, tmpDir);

const url = 'http://localhost:' + port;

const jsonClient = restify.createJsonClient({
  url: url,
  version: '~1.0'
});


describe('management', function () {
  /**
   * Start and stop the server before and after the test suites
   */
  before(function () {
    server.listen()
  });

  after(function () {
    server.close();
  });

  describe('#heartbeat()', function () {
    it('should get without error', function (done) {
      jsonClient.get('/v1/heartbeat', function (err, req, res, data) {
        if (err) {
          throw new Error(err);
        }

        should(res).not.null();
        should(res.statusCode).not.null();
        should(res.statusCode).not.undefined();
        res.statusCode.should.equal(200);


        should(data).not.null();
        should(data.status).not.undefined();
        data.status.should.equal('ok');


        done();

      });
    });
  });

  describe('#upload()', function () {
    it('happy path with zip', function (done) {

      const client = new Client('localhost', port);

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/echo-test.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err);
        }
        should(response).not.null();
        should(response.statusCode).not.null();
        should(response.statusCode).not.undefined();
        response.statusCode.should.equal(200);


        const json = JSON.parse(bodyBuffer);

        should(json).not.null();
        should(json.endpoint).not.undefined();
        json.endpoint.should.equal('http://endpointyouhit:8080');

        done();

      });


    });

    it('valid zip no package.json ', function (done) {

      const client = new Client('localhost', port);

      client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/text-file.zip', function (err, response, bodyBuffer) {
        if (err) {
          throw new Error(err);
        }

        should(response).not.null();
        should(response.statusCode).not.null();
        should(response.statusCode).not.undefined();
        response.statusCode.should.equal(400);


        response.statusMessage.should.equal('no package.json defined');

        done();

      });


    });


    it('not a valid zip ', function (done) {

          const client = new Client('localhost', port);

          client.putZipFile('testOrg', 'testEnv', 'testApp', 1, 'test/assets/not-a-zip.zip', function (err, response, bodyBuffer) {
            if (err) {
              throw new Error(err);
            }

            should(response).not.null();
            should(response.statusCode).not.null();
            should(response.statusCode).not.undefined();
            response.statusCode.should.equal(400);


            response.statusMessage.should.equal('not a zip file');

            done();

          });


        });
  });
});


/**
 *
 * @constructor
 */
const Client = function (host, port) {
  this.host = host;
  this.port = port;
}

/**
 * Post the zip file.  Callback is of the type
 *
 * function(err, response, bodyBuffer)
 */
Client.prototype.putZipFile = function (orgName, envName, appName, revision, zipFilePath, cb) {

  fs.existsSync(zipFilePath).should.equal(true);

  const readStream = fs.createReadStream(zipFilePath);

  const headers = {};
  headers['x-apigee-script-container-rev'] = revision;

  var options = {
    host: 'localhost'
    , port: port
    , path: '/v1/deploy/' + orgName + '/' + envName + '/' + appName
    , method: 'PUT'
    , headers: headers
  };

  var req = http.request(options, function (res) {
    //this feels a bit backwards, but these are evaluated AFTER the read stream has closed

    var buffer = '';

    //pipe body to a buffer
    res.on('data', function (data) {
      buffer += data;
    });

    res.on('end', function () {

      cb(null, res, buffer);

    });

  });

  req.on('error', function (err) {
    if (err) {
      cb(err, null, null);
    }

  });

  //pipe the readstream into the request
  readStream.pipe(req);


  /**
   * Close the request on the close of the read stream
   */
  readStream.on('close', function () {
    req.end();
    console.log('I finished.');
  });


  //note that if we end up with larger files, we may want to support the continue, much as S3 does
  //https://nodejs.org/api/http.html#http_event_continue


};

