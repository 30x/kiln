'use strict'

/**
 * Class to handle temporary file I/O and stream limiting for the file upload
 */
const fs = require('fs')
const limitStream = require('size-limit-stream')
const pumpify = require('pumpify')
const unzip = require('unzip')
const rimraf = require('rimraf')
const tarzan = require('tarzan');

const PACKAGE_JSON = 'package.json'


//TODO REFACTOR this into a utility module without state.  State is now externalized in appInfo
function Io() {

}

module.exports = Io


/**
 * Generate a temp file stream with a max limit imposed and return it
 *
 * @param appInfo The ioInfo object
 * @param cb, the callback with an err argument
 */
Io.prototype.createZipFileStream = function (appInfo, cb) {
  const fileWriteStream = fs.createWriteStream(appInfo.zipFile)

  const limiter = limitStream(appInfo.maxFileSize)

  var combined = pumpify(limiter, fileWriteStream)
  combined.on('error', cb)

  return combined
}

/**
 * Unlink the temp files and tars.  Ignores missing files
 * @param ioInfo the IoInfo object
 */
Io.prototype.cleanup = function (ioInfo) {
  /**
   * Unlink the file. If it fails, log and ignore
   */
  fs.unlink(ioInfo.zipFile, function (err) {
    if (err) {
      console.error('Unable to delete temp file at %s', tempFileName)
    }

  })

  fs.unlink(ioInfo.tarFile, function (err) {
    if (err) {
      console.error('Unable to delete temp file at %s', tempFileName)
    }

  })

  rimraf(ioInfo.outputDir, {}, function (err) {
    if (err) {
      console.error('Unable to remove directory at %s', outputDir)
      console.error(err)
    }
  })
}


/**
 * Validate the zip file.  Invoke the callback with any errors
 * @param appInfo
 * @param cb A callback of the form (err)
 */
Io.prototype.extractZip = function (appInfo, cb) {
  const stream = fs.createReadStream(appInfo.zipFile).pipe(unzip.Extract({path: appInfo.outputDir}))

  stream.on('close', cb)

  stream.on('error', cb)
}

/**
 * Validate our output zip. contains a valid package.json.  Does not perform any other output
 * @param appInfo
 * @param cb A function that takes an err. If no err is present, the validate is successful
 */
Io.prototype.validateZip = function (appInfo, cb) {

  const packageJson = appInfo.outputDir + '/' + PACKAGE_JSON

  fs.readFile(packageJson, 'utf8', function (err, data) {
    //couldn't read the file
    if (err) {
      console.log("Unable to read package.json: %s", err)
      return cb(new Error("package.json could not be read.  Ensure it is in your upload."))
    }

    //parse the file
    const parsed = JSON.parse(data)

    /**
     * Check the fields are present we need to run the application
     */
    if (!parsed.scripts) {
      return cb(new Error("scripts.run is required in package.json."))
    }


    if (!parsed.scripts.start) {
      return cb(new Error("scripts.start is required in package.json."))
    }


    cb()

  })


}


/**
 * Copy the docker file into the output directory
 * @param appInfo
 * @param cb
 */
Io.prototype.copyDockerfile = function (appInfo, cb) {
  fs.createReadStream('assets/Dockerfile').pipe(fs.createWriteStream(appInfo.outputDir + '/Dockerfile'))
    .on('error', function (err) {
      console.error(err)

      return cb(err)

    })
    .on('close', function () {
      console.log('Finished copying docker file to %s', appInfo.outputDir)
      return cb(null)
    })
}

/**
 * Create a tar file
 * @param appInfo
 * @param cb the callback of the form (err).
 */
Io.prototype.createTarFile = function (appInfo, cb) {
  tarzan({directory: appInfo.outputDir}).pipe(fs.createWriteStream(appInfo.tarFile))

      .on('error', function (err) {
        console.error(err)

        return cb(err)

      })
      .on('close', function () {
        console.log('Finished building tar at %s', appInfo.tarFile)
        return cb(null)
      })

  //tarPack.pack(appInfo.outputDir).pipe(fs.createWriteStream(appInfo.tarFile))
  //
  //  .on('error', function (err) {
  //    console.error(err)
  //
  //    return cb(err)
  //
  //  })
  //  .on('close', function () {
  //    console.log('Finished building tar at %s', appInfo.tarFile)
  //    return cb(null)
  //  })
}

Io.prototype.getTarFileStream = function (appInfo){
  return fs.createReadStream(appInfo.tarFile)
}


