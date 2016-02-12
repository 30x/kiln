'use strict'


const Dockerode = require('dockerode')
const fs = require('fs')
const Io = require('./io.js')

/**
 * Docker api client.  Intialize with the TCP ip and port
 *
 * TODO: Create a constructor that can take in dockerode socker for local execution.
 * this reads from environment variables of the following
 *
 * DOCKER_HOST The host path to connect to
 * DOCKER_TLS_VERIFY.  If 1, then DOCKER_CERT_PATH will be read to sign requests
 *
 *
 * Here is a link to some issue when running docker-machine on your local host and testing with curl.
 * https://github.com/docker/docker/issues/16107
 *
 * You have to test in the following way.
 *
 *
 * $openssl pkcs12 -export -inkey $DOCKER_CERT_PATH/key.pem -in $DOCKER_CERT_PATH/cert.pem -name test-curl-client-side
 * -out test-curl.p12 -password pass:mysecret
 *
 * $/usr/local/Cellar/curl/7.43.0/bin/curl -k --cert test-curl.p12:mysecret  https://$(docker-machine ip <your machine
 * name>):2376/info
 * @constructor
 */
function Docker() {

  /**
   * Note this has only been tested with docker machine.  Will need tested when using local unix socket
   */
  const dockerHost = process.env.DOCKER_HOST

  if (!dockerHost) {
    throw new Error('You must set the DOCKER_HOST environment variable')
  }

  const tlsVerify = process.env.DOCKER_TLS_VERIFY

  const dockerUrl = new DockerUrl(dockerHost)


  if (tlsVerify == 1 && dockerUrl.matches && (dockerUrl.protocol === 'tcp' || dockerUrl.protocol === 'http')) {


    const certDir = process.env.DOCKER_CERT_PATH

    if (!certDir) {
      throw new Error('When using DOCKER_TLS_VERIFY=1 you must specify the property DOCKER_CERT_PATH for certificates')
    }

    this.docker = new Dockerode({
      host: dockerUrl.hostName,
      port: dockerUrl.port,
      checkServerIdentity: false,
      ca: fs.readFileSync(certDir + '/ca.pem'),
      cert: fs.readFileSync(certDir + '/cert.pem'),
      key: fs.readFileSync(certDir + '/key.pem')
    })
  }


  /**
   * Nothing created the client, blow up
   */
  if (!this.docker) {
    throw new Error('Could not interpret environment variables to create docker client')
  }

  ///**
  // * Validate it works, the process should die if it fails on error
  // */
  //this.docker.info(function (err, data) {
  //  if (err) {
  //    throw err
  //  }
  //
  //  console.log('docker info %s', data)
  //})

  this.io = new Io()


}


module.exports = Docker


/**
 * Create a container.  Invokes the callback with an err or the container Id
 * @param appInfo The application's info object
 * @param cb The callback of the form (err, containerId)
 */
Docker.prototype.createContainer = function (appInfo, cb) {

  //TODO, build image in the following

  //
  //1) copy docker file
  //2) tar it up
  //3) send it off

  const docker = this.docker
  const io = this.io;

  io.createTarFile(appInfo, function (err) {
    if (err) {
      return cb(err)
    }

    //no error creating the tar, build the image


    const fileStream = io.getTarFileStream(appInfo)

    docker.buildImage(fileStream, {t: appInfo.tagName}, function (err, stream) {


      if (err) {
        throw err
      }

      stream.pipe(process.stdout, {
        end: true
      });

      stream.on('end', function () {
        done();

        //when we're done, get the container id
        //docker.

        const containerId = response.containerId

        return cb(null, containerId)
      });


    })
  })

}

/**
 * Pulls the image.  When complete, calls the callback.
 * @param repoTag
 * @param cb
 */
Docker.prototype.initialize = function (repoTag, cb) {


  this.docker.pull(repoTag, function (err, stream) {
    if (err) return cb(err);
    stream.pipe(process.stdout);
    stream.once('end', cb);
  });

}


/**
 * Creates a docker URL for parsing
 * @param url
 * @constructor
 */
function DockerUrl(url) {
  const tcpOrHttp = new RegExp(/(\w+):\/\/(.*):(\d+)/)

  const parts = url.match(tcpOrHttp)

  this.matches = parts.length === 4

  if (!this.matches) {
    return
  }

  this.protocol = parts[1]
  this.hostName = parts[2]
  this.port = parts[3]
}

