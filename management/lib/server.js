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

    res.send(util.format('logs go here for %s', mpName));
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

    const revision = req.header(revisionHeader);

    if (!orgName) {
      return next(new restify.BadRequestError({statusCode: 400, message: 'You must specify an org name'}));
    }

    if (!envName) {
      return next(new restify.BadRequestError({statusCode: 400, message: 'You must specify an env name'}));
    }

    if (!appName) {
      return next(new restify.BadRequestError({statusCode: 400, message: 'You must specify an app name'}));
    }

    if (!revision) {
      return next(new restify.BadRequestError({
        statusCode: 400,
        message: 'You must specify a version in x-apigee-script-container-rev header'
      }));
    }

    //limit our file size input

    const tempFileName = io.createFileName(orgName, envName, appName, revision);

    const fileWriteStream = io.createInputStream(tempFileName, function (err) {

      if (err) {
        //return an error if we exceed the file size
        io.unlinkTempFile(tempFileName);
        console.error('Unable to accept zip file.  %s', err);

        return next(new restify.InternalServerError({statusCode: 500, message: 'Unable to accept zip file'}));
      }
    });

    //pipe the request body to the file
    const fileStream = req.pipe(fileWriteStream);

    fileStream.on('error', function (err) {
      if (err) {

        //return an error
        console.error('Unable to create file stream %s', err);
        io.unlinkTempFile(tempFileName);
        return next(new restify.InternalServerError({statusCode: 500, message: 'Unable to accept zip file'}));
      }
    });

    //once we're done writing the stream, render a response
    fileStream.on('finish', function () {
      const outputDirName = io.createOutputDirName(orgName, envName, appName, revision);

      //extract the zip file and validate it
      io.extractZip(tempFileName, outputDirName, function (err) {
        if (err) {

          console.error('Unable to extract zip file %s', err);



          io.unlinkTempFile(tempFileName);
          io.deleteExtractedZip(outputDirName);

          return next(new restify.BadRequestError({statusCode: 400, message: 'Unable to extract zip file.  Ensure you have a valid zip file'}));
        }

        //validate the zip file
        io.validateZip(outputDirName, function (err) {
          if (err) {

            console.error('Unable to validate zip file %s', err);

            io.unlinkTempFile(tempFileName);
            io.deleteExtractedZip(outputDirName);

            return next(new restify.BadRequestError({statusCode: 400, message: 'Unable to validate node application. ' + err.message} ));
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
