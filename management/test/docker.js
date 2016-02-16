'use strict'

const should = require('should')
const uuid = require('uuid')
const Docker = require('../lib/docker.js')
const AppInfo = require('../lib/appinfo.js')
const Io = require('../lib/io.js')
const testConstants = require('./testconfig.js')
const numGenerator = require('./numbergen.js')

const restify = require('restify')


describe('docker', function () {

  var docker
  var appInfo
  var io

  /**
   * Start and stop the server before and after the test suites
   */
  before(function (done) {

    appInfo = new AppInfo(testConstants.tmpDir, "orgName", "envName", "appName" + numGenerator.randomInt(), 1, testConstants.maxFileSize)
    docker = new Docker(testConstants.dockerUrl)
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
      docker.initialize(testConstants.defaultImage, function (err) {

        if (err) {
          done(err)
        }

        done()
      })

    })

    it('create container', function (done) {

      //creating a container can be slow the first time it runs
      this.timeout(120000)

      docker.createContainer(appInfo, function (err, dockerInfo) {

        if (err) {
          return done(err)
        }

        should(dockerInfo).not.null()
        should(dockerInfo).not.undefined()

        dockerInfo.containerName.should.equal(appInfo.containerName)

        dockerInfo.revision.should.equal(appInfo.revision)

        dockerInfo.remoteTag.should.equal(appInfo.remoteTag)

        dockerInfo.remoteContainer.should.equal(appInfo.remoteContainer)

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


        //now the container is created, push it

        docker.tagAndPush(appInfo, function (err, dockerInfo) {
          if (err) {
            throw err
          }

          should(dockerInfo).not.null()
          should(dockerInfo).not.undefined()

          dockerInfo.containerName.should.equal(appInfo.containerName)

          dockerInfo.revision.should.equal(appInfo.revision)

          dockerInfo.remoteTag.should.equal(appInfo.remoteTag)

          dockerInfo.remoteContainer.should.equal(appInfo.remoteContainer)

          //TODO connect to remote repository and validate image.  Validated by hand

          //Get all images curl -X GET localhost:5000/v2/_catalog

          //get tags for an image   curl -X GET localhost:5000/v2/<name>/tags/list
          // curl -X GET localhost:5000/v2/orgname_envname/appname14040/tags/list

          const jsonClient = restify.createJsonClient({
            url: testConstants.dockerUrl,
            version: '~1.0'
          })


          //now get them from the remote repo, make sure it actually worked

          jsonClient.get('/v2/_catalog', function (err, req, res, data) {
            if (err) {
              throw new Error(err)
            }

            should(res).not.null()
            should(res.statusCode).not.null()
            should(res.statusCode).not.undefined()
            res.statusCode.should.equal(200)


            should(data).not.null()
            should(data.repositories).not.undefined()
            //should exist above a -1
            data.repositories.indexOf(appInfo.containerName).should.above(-1)


            //now get the specific version and be sure it's there

            const url = '/v2/' + appInfo.containerName + "/tags/list"

            jsonClient.get(url, function (err, req, res, data) {
              if (err) {
                throw new Error(err)
              }

              should(res).not.null()
              should(res.statusCode).not.null()
              should(res.statusCode).not.undefined()
              res.statusCode.should.equal(200)


              should(data).not.null()
              should(data.tags).not.undefined()
              //should exist above a -1
              data.tags.indexOf(appInfo.revision).should.above(-1)


              done()

            })
          })

        })
      })

    })


  })
})


