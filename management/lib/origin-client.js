'use strict';

const fs = require('fs')
const helpers = require('./helpers')

/**
 * The Origin configuration.
 *
 * @typedef {object} OriginClientConfig
 *
 * @property {string} apiBase - The fully qualified URL to the Origin API base
 * @property {string} apiToken - The API token to use for making Origin API calls
 */

/**
 * Creates an Origin client.
 *
 * @param {~OriginClientConfig} [conf] - The client configuration *(If omitted, default values will be used that assume
 * the Origin client is run within a Docker container, on Origin.)*
 *
 * @constructor
 */
function Client (conf) {
  this.conf = conf || {}

  // Attempt to fill in defaults if/when necessary
  if (conf) {
    if (!conf.apiBase)
      this.conf.apiBase = 'https://openshift.default.svc.cluster.local/oapi/v1'

    if (!conf.apiToken)
      this.conf.apiToken = fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf-8')
  }
}

/**
 * Returns all projects for the cluster.  *(Must be ran as a cluster admin)*
 *
 * @param {function} done - The error-first callback
 *
 * @see https://docs.openshift.org/latest/rest_api/openshift_v1.html#list-objects-of-kind-project
 */
Client.prototype.getProjects = function (done) {
  helpers.makeApiRequest(this.conf, {
    path: '/projects'
  }, (err, res) => {
    if (err)
      done(err)
    else
      done(null, JSON.parse(res.text).items)
  })
}

module.exports = Client
