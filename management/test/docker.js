'use strict'

const should = require('should')
const uuid = require('uuid')
const Docker = require('../lib/docker.js')
const AppInfo = require('../lib/appinfo.js')
const Io = require('../lib/io.js')


describe('docker', function () {

  var docker
  var appInfo
  var io

  /**
   * Start and stop the server before and after the test suites
   */
  before(function (done) {

    appInfo = new AppInfo('/tmp', "orgName", "envName", "appName", 1, 104857600)
    docker = new Docker()
    io = new Io()

    //override the zip file location so we don't have to copy the file

    appInfo.zipFile = 'test/assets/echo-test.zip'

    io.extractZip(appInfo, function (err) {
      if (err) {
        return done(err)
      }

      //copy the docker file to the output for packing
      io.copyDockerfile(appInfo, function (err) {

        if (err) {
          return done(err)
        }
        done()
      })


    })


  })

  after(function () {

  })

  describe('#docker()', function () {
    it('init node container', function (done) {

      //downloading the initial onbuild can be slow

      this.timeout(60000);
      docker.initialize('node:4.3.0-onbuild', function (err) {

        if (err) {
          throw err
        }

        done()
      })

    })

    it('create container', function (done) {

      docker.createContainer(appInfo, function (err, dockerId) {

        if (err) {
          throw err
        }

        should(dockerId).not.null()
        should(dockerId).not.undefined()

        done()
      })

    })

    it('tag container', function (done) {
      done()
    })

    it('push container', function (done) {
      done()
    })
  })
})


