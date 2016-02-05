'use strict'
const should = require('should');
const restify = require('restify');
const fs = require('fs');
const http = require('http');
const server = require('../lib/server.js');

const port = 8080;
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
        server.listen(port)
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


            const testPath = 'test/assets/test.zip';
            fs.existsSync(testPath).should.equal(true);


            const readStream = fs.createReadStream(testPath);


            var options = {
                host: 'localhost'
                , port: port
                , path: '/v1/deploy/testvalue'
                , method: 'PUT'
            };

            var req = http.request(options, function (res) {
                //this feels a bit backwards, but these are evaluated AFTER the read stream has closed

                var buffer = '';

                //pipe body to a buffer
                res.on('data', function(data){
                    buffer+= data;
                });

                res.on('end', function () {

                    should(res).not.null();
                    should(res.statusCode).not.null();
                    should(res.statusCode).not.undefined();
                    res.statusCode.should.equal(200);


                    const json = JSON.parse(buffer);

                    should(json).not.null();
                    should(json.endpoint).not.undefined();
                    json.endpoint.should.equal('http://endpointyouhit:8080');


                    done();
                });

            });

            req.on('error', function (err) {
                if (err) {
                    throw new Error(err);
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

        });
    });
});
