'use strict'

const should = require('should')
const uuid = require('uuid')
const Docker = require('../lib/docker.js')
const AppInfo = require('../lib/appinfo.js')
const Io = require('../lib/io.js')

function randomInt () {
    return Math.floor(Math.random() * (100000 - 0) + 0);
}


describe('docker', function () {

  var docker
  var appInfo
  var io

  /**
   * Start and stop the server before and after the test suites
   */
  before(function (done) {

    appInfo = new AppInfo('/tmp', "orgName", "envName", "appName"+randomInt(), 1, 104857600)
    docker = new Docker('localhost:5000')
    io = new Io()

    //override the zip file location so we don't have to copy the file

    appInfo.zipFile = 'test/assets/echo-test.zip'

    io.extractZip(appInfo, function (err) {
      if (err) {
        return done(err)
      }


      done()


    })


  })

  after(function () {

  })

  describe('#docker()', function () {
    it('init node container', function (done) {

      //downloading the initial onbuild can be slow

      this.timeout(120000);
      docker.initialize('node:4.3.0-onbuild', function (err) {

        if (err) {
          done(err)
        }

        done()
      })

    })

    it('create container', function (done) {

      //creating a container can be slow the first time it runs
      this.timeout(120000)

      docker.createContainer(appInfo, function (err, dockerId) {

        if (err) {
          return done(err)
        }

        should(dockerId).not.null()
        should(dockerId).not.undefined()

        dockerId.should.equal(appInfo.tagName)

        done()
      })

    })

    it('tag and push container', function (done) {
      //create the container first
      docker.createContainer(appInfo, function (err, dockerId) {

        if (err) {
          done(err)
        }

        should(dockerId).not.null()
        should(dockerId).not.undefined()

        dockerId.should.equal(appInfo.tagName)

        //now the container is created, push it

        docker.tagAndPush(appInfo, function (err, result) {
          if (err) {
            throw err
          }

          result.should.equal(dockerId)

          //TODO connect to remote repository and validate image.  Validated by hand
          done()
        })
      })

    })



  })
})


