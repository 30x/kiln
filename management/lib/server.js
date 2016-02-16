'use strict'
/**
 * Imports
 *
 */
const restify = require('restify')
const util = require('util')
const Io = require('./io.js')
const AppInfo = require('./appinfo.js')


/**
 * Constants
 *
 */
const revisionHeader = 'x-apigee-script-container-rev'


/**
 * Create a new instance of the server.
 * @param port The port to listen on
 * @param tmpDir The temporary directory
 * @param maxFileSize The max file size to allow during upload
 * @param docker an instance of the docker object for communicating with the docker api
 * @constructor
 */
function Server(port, tmpDir, maxFileSize, dockerInstance) {
  this.port = Number(port)
  const docker = dockerInstance
  const io = new Io()

  this.server = restify.createServer({
    name: 'shipyard',
    version: '1.0.0'
  })

  /**
   * configure the server options
   */
  this.server.use(restify.acceptParser(this.server.acceptable))
  /**
   * Parse query params
   */
  this.server.use(restify.queryParser())
  /**
   * only parse json body. TODO, check this for security
   */
  this.server.use(restify.jsonBodyParser())

  /**
   * Heartbeat for system liveliness
   * TODO integrate with origin/k8s call to ensure e2e communication
   */
  this.server.get('/v1/heartbeat', function (req, res, next) {
    res.send({status: 'ok'})
    return next()
  })


  /**
   * Get the logs
   */
  this.server.get('/v1/logs/:mpname', function (req, res, next) {
    const mpName = req.mpname

    res.send(util.format('logs go here for %s', mpName))
    return next()
  })

  /**
   * The endpoint where the Zip file is posted to
   */
  this.server.put('/v1/buildnodejs/:org/:env/:app', function (req, res, next) {


      //split for clarity
      const orgName = req.params.org
      const envName = req.params.env
      const appName = req.params.app

      const revision = req.header(revisionHeader)

      if (!orgName) {
        return next(new restify.errors.BadRequestError({statusCode: 400, message: 'You must specify an org name'}))
      }

      if (!envName) {
        return next(new restify.errors.BadRequestError({statusCode: 400, message: 'You must specify an env name'}))
      }

      if (!appName) {
        return next(new restify.errors.BadRequestError({statusCode: 400, message: 'You must specify an app name'}))
      }

      if (!revision) {
        return next(new restify.errors.BadRequestError({
          statusCode: 400,
          message: 'You must specify a version in x-apigee-script-container-rev header'
        }))
      }

      const appInfo = new AppInfo(tmpDir, orgName, envName, appName, revision, maxFileSize)

      //limit our file size input

      const fileWriteStream = io.createZipFileStream(appInfo, function (err) {

        if (err) {
          //return an error if we exceed the file size
          io.cleanup(appInfo)
          console.error('Unable to accept zip file.  %s', err)

          return next(new restify.errors.errorsHttpError({statusCode: 500, message: 'Unable to accept zip file'}))
        }
      })

      //pipe the request body to the file
      const fileStream = req.pipe(fileWriteStream)

      fileStream.on('error', function (err) {
        if (err) {

          //return an error
          console.error('Unable to create file stream %s', err)
          io.cleanup(appInfo)
          return next(new restify.errors.InternalServerError({statusCode: 500, message: 'Unable to accept zip file'}))
        }
      })

      //once we're done writing the stream, render a response
      fileStream.on('finish', function () {
        //extract the zip file and validate it
        io.extractZip(appInfo, function (err) {
          if (err) {

            console.error('Unable to extract zip file %s', err)

            io.cleanup(appInfo)
            return next(new restify.errors.BadRequestError({
              statusCode: 400,
              message: 'Unable to extract zip file.  Ensure you have a valid zip file.'
            }))
          }

          //validate the zip file
          io.validateZip(appInfo, function (err) {
            if (err) {

              console.error('Unable to validate zip file %s', err)

              io.cleanup(appInfo)

              return next(new restify.errors.BadRequestError({
                statusCode: 400,
                message: 'Unable to validate node application. ' + err.message
              }))
            }


            docker.createContainer(appInfo, function (err, containerId) {


              //we failed to create, cleanup and fail
              if (err) {
                io.cleanup(appInfo)
                return next(new restify.errors.BadRequestError({
                  statusCode: 400,
                  message: 'Unable to create docker container. ' + err.message
                }))
              }


              //tag the image and push it to the repo
              docker.tagAndPush(appInfo, function (err, containerId) {

                //cleanup regardless of success or failure, we need to either way
                io.cleanup(appInfo)

                if (err) {
                  return next(new restify.errors.BadRequestError({
                    statusCode: 400,
                    message: 'Unable to tag and push the docker container. ' + err.message
                  }))
                }


                //send back the endpoint the caller should hit for the deployed application
                res.send({endpoint: 'http://endpointyouhit:8080', containerId: containerId})

                return next()

              })


            })


          })

        })


      })


    }
  )

}

module.exports = Server


Server.prototype.listen = function () {
  const that = this

  that.server.listen(this.port, function () {
    console.log('%s listening at %s', that.server.name, that.server.url)
  })
}


Server.prototype.close = function () {
  this.server.close()
}
