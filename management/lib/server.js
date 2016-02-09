'use strict'
/**
 * Imports
 *
 */
const restify = require('restify');
const util = require('util');
const Io = require('./io.js');


/**
 * Constants
 *
 */
const revisionHeader = 'x-apigee-script-container-rev';


function Server(port, tmpDir, maxFileSize) {
  this.port = Number(port);
  const io = new Io(tmpDir, maxFileSize);

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
  this.server.put('/v1/deploy/:org/:env/:app', function (req, res, next) {


    //split for clarity
    const orgName = req.params.org;
    const envName = req.params.env;
    const appName = req.params.app;

    const revision =  req.header(revisionHeader);

    if (!orgName) {
      return next(new restify.errors.BadRequestError("You must specify an org name"));
    }

    if (!envName) {
      return next(new restify.errors.BadRequestError("You must specify an env name"));
    }

    if (!appName) {
      return next(new restify.errors.BadRequestError("You must specify an app name"));
    }

    if(!revision){
      return next(new restify.errors.BadRequestError("You must specify a version in x-apigee-script-container-rev header"));
    }

    //limit our file size input

    const tempFileName = io.createFileName(orgName, envName, appName, revision);

    const fileWriteStream = io.createInputStream(tempFileName, function (err) {
      //return an error if we exceed the file size
      io.unlinkTempFile(tempFileName);
      return next(new restify.errors.InternalServerError(err));
    });

    //pipe the request body to the file
    const fileStream = req.pipe(fileWriteStream);

    fileStream.on('error', function (err) {
      //return an error
      io.unlinkTempFile(tempFileName);
      return next(new restify.errors.InternalServerError(err));
    });

    //once we're done writing the stream, render a response
    fileStream.on('finish', function () {
      const outputDirName = io.createOutputDirName(orgName, envName, appName, revision);

      //extract the zip file and validate it
      io.extractZip(tempFileName, outputDirName, function(err){
        if(err){
          return next(new restify.errors.BadRequestError(err));
        }

        //validate the zip file
        io.validateZip(outputDirName, function(err){
          if(err){
            return next(new restify.errors.BadRequestError(err));
          }

          //the json is valid, TODO deploy here

          //ignore errors on delete, if everything else is successful, we just want to log them.

          io.unlinkTempFile(tempFileName);
          io.deleteExtractedZip(outputDirName);


          //send back the endpoint the caller should hit for the deployed application
          res.send({endpoint: 'http://endpointyouhit:8080'});

          return next();


        });

      });


    });


  });

}

module.exports = Server;


Server.prototype.listen = function () {
  const that = this;

  that.server.listen(this.port, function () {
    console.log('%s listening at %s', that.server.name, that.server.url);
  });
};


Server.prototype.close = function () {
  this.server.close();
};
