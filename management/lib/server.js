'use strict'
const restify = require('restify');
const fs = require('fs');
const util = require('util');
const uuid = require('uuid');
const temp = require('temp').track();
const limitStream = require('size-limit-stream');


function Server(port, tmpDir) {
  this.port = port;
  this.tmpDir = tmpDir;

  this.server = restify.createServer({
    name: 'shipyard',
    version: '1.0.0'
  });

  /**
   * configure the server options
   */
  this.server.use(restify.acceptParser(this.server.acceptable));
  /**
   * Parse query params
   */
  this.server.use(restify.queryParser());
  /**
   * only parse json body. TODO, check this for security
   */
  this.server.use(restify.jsonBodyParser());

  /**
   * Heartbeat for system liveliness
   * TODO integrate with origin/k8s call to ensure e2e communication
   */
  this.server.get('/v1/heartbeat', function (req, res, next) {
    res.send({status: 'ok'});
    return next();
  });


  /**
   * Get the logs
   */
  this.server.get('/v1/logs/:mpname', function (req, res, next) {
    const mpName = req.mpname;

    res.send(util.format("logs go here for %s", mpName));
    return next();
  });

  /**
   * The endpoint where the Zip file is posted to
   */
  this.server.put('/v1/deploy/:name', function (req, res, next) {


    //split for clarity
    const assetName = req.params.name;


    const tempFileName = tmpDir + '/' + assetName +'-'+ uuid.v1() +'.zip';

    const fileWriteStream = fs.createWriteStream(tempFileName);

    const fileStream = req.pipe(fileWriteStream);

    fileStream.on('error', function (err) {
      //return an error
      return next(new restify.errors.InternalServerError(err));
    });

    //once we're done writing the stream, render a response
    fileStream.on('finish', function () {
      //TODO: Real work here

      /**
       * Unlink the file. If it fails, log and ignore
       */
      fs.unlink(tempFileName, function(err){
        if(err){
          console.log('Unable to delete temp file at %s', tempFileName);
        }

      });
      //send back the endpoint the caller should hit for the deployed application
      res.send({endpoint: 'http://endpointyouhit:8080'});

      return next();
    });



  });

}

module.exports = Server;


Server.prototype.listen = function () {
  const serverPointer = this.server;

  serverPointer.listen(this.port, function () {
    console.log('%s listening at %s', serverPointer.name, serverPointer.url);
  });
};


Server.prototype.close = function () {
  this.server.close();
};
