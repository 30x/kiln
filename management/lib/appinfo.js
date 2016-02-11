'use strict'
const uuid = require('uuid')

/**
 * @param tmpDir The temporary directory
 * @param orgName The org name
 * @param envName The environment name
 * @param appName The app name
 * @param revision The revision of the deployment
 * @param maxFileSize the maximum file size
 * @constructor Returns an instance with 3 properties.  outputDir, zipFile, and tarFile
 */
function AppInfo(tmpDir, orgName, envName, appName, revision, maxFileSize) {

  //we throw in a time uuid just in case we get double PUT.  Both requests can process, and first to finish will win
  //without creating a race condition on the file stream.

  this.orgName = orgName
  this.envName = envName
  this.appname = appName
  this.revision = revision
  this.maxFileSize = maxFileSize

  //this.tagName = orgName  + envName  + appName + ':' + revision
  this.tagName = (orgName + '_' + envName + '_' + appName + ':' + revision).toLowerCase()


  this.outputDir = tmpDir + '/' + orgName + '_' + envName + '_' + appName + '_' + revision + '_' + uuid.v1()

  this.zipFile = this.outputDir + ".zip"

  this.tarFile = this.outputDir + ".tar"
}

module.exports = AppInfo
