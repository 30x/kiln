'use strict'
const should = require('should');
const restify = require('restify');
const server = require('../lib/server.js');
 
const port = 8080;

const client = restify.createJsonClient({
  url: 'http://localhost:'+port,
  version: '~1.0'
});

describe('management', function() {
   /**
    * Start and stop the server before and after the test suites
    */
   before(function() {
    server.listen(port)
  });

  after(function() {
    server.close();
  });  
    
  describe('#heartbeat()', function() {
    it('should get without error', function(done) {
      client.get('/v1/management/heartbeat', function(err, req, res, data) {
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
  });