'use strict';

/**
 * Class to handle temporary file I/O and stream limiting for the file upload
 */
const uuid = require('uuid');
const fs = require('fs');
const limitStream = require('size-limit-stream');
const pumpify = require('pumpify');
const unzip = require('unzip');
const rimraf = require('rimraf');

const PACKAGE_JSON = 'package.json';

/**
 *
 * @param tmpDir
 * @constructor
 */
function Io(tmpDir, maxFileSize) {
  this.tmpDir = tmpDir;
  this.maxFileSize = Number(maxFileSize);
}


module.exports = Io;


/**
 * Create a file name for a temp file on the file system
 * @param orgName The org name
 * @param envName The environment name
 * @param appName The app name
 * @param revision The revision of the deployment
 * @returns {string}
 */
Io.prototype.createFileName = function (orgName, envName, appName, revision) {
  //we throw in a time uuid just in case we get double PUT.  Both requests can process, and first to finish will win
  //without creating a race condition on the file stream.
  return this.createOutputDirName(orgName, envName, appName, revision) + ".zip";
};

/**
 * Create a file name for a temp file on the file system
 * @param orgName The org name
 * @param envName The environment name
 * @param appName The app name
 * @param revision The revision of the deployment
 * @returns {string}
 */
Io.prototype.createOutputDirName = function (orgName, envName, appName, revision) {
  //we throw in a time uuid just in case we get double PUT.  Both requests can process, and first to finish will win
  //without creating a race condition on the file stream.
  return this.tmpDir + '/' + orgName + '-' + envName + '-' + appName + '-' + revision + '-' + uuid.v1();
};
/**
 * Generate a temp file stream with a max limit imposed and return it
 *
 * @param tempFileName
 */
Io.prototype.createInputStream = function (tempFileName, cb) {
  const fileWriteStream = fs.createWriteStream(tempFileName);

  const limiter = limitStream(this.maxFileSize);

  var combined = pumpify(limiter, fileWriteStream);
  combined.on('error', cb);

  return combined;
};

/**
 * Unlink the temp file
 * @param tempFileName
 */
Io.prototype.unlinkTempFile = function (tempFileName) {
  /**
   * Unlink the file. If it fails, log and ignore
   */
  fs.unlink(tempFileName, function (err) {
    if (err) {
      console.error('Unable to delete temp file at %s', tempFileName);
    }

  });
};

/**
 * Validate the zip file.  Invoke the callback with any errors
 * @param tempFileName
 * @param cb
 */
Io.prototype.extractZip = function (tempFileName, outputDir, cb) {
  const stream = fs.createReadStream(tempFileName).pipe(unzip.Extract({path: outputDir}));

  stream.on('close', cb);

  stream.on('error', cb);
};


/**
 * Delete the extracted zip
 * @param outputDir
 * @param cb
 */
Io.prototype.deleteExtractedZip = function(outputDir){
  rimraf(outputDir, {}, function(err){
    if(err){
      console.error('Unable to remove directory at %s', outputDir);
      console.error(err);
    }
  });
};
/**
 * Validate our output zip. contains a valid package.json.  Does not perform any other output
 * @param outputDir
 * @param cb A function that takes an err. If no err is present, the validate is successful
 */
Io.prototype.validateZip = function (outputDir, cb) {

  const packageJson = outputDir + '/' + PACKAGE_JSON;

  fs.readFile(packageJson, 'utf8', function (err, data) {
    //couldn't read the file
    if (err) {
      console.log("Unable to read package.json: %s", err);
      return cb(new Error("package.json could not be read.  Ensure it is in your upload"));
    }

    //parse the file
    const parsed = JSON.parse(data);

    /**
     * Check the fields are present we need to run the application
     */
    if (!parsed.scripts) {
      return cb("scripts.run is required in package.json");
    }


    if (!parsed.scripts.start) {
      return cb("scripts.start is required in package.json");
    }



    cb();

  });


};
