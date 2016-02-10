'use strict';

const fs = require('fs')
const helpers = require('./helpers')

/**
 * The Kubernetes configuration.
 *
 * @typedef {object} KubernetesClientConfig
 *
 * @property {string} apiBase - The fully qualified URL to the Kubernetes API base
 * @property {string} apiToken - The API token to use for making Kubernetes API calls
 */

/**
 * Creates an Kubernetes client.
 *
 * @param {~KubernetesClientConfig} [conf] - The client configuration *(If omitted, default values will be used that assume
 * the Kubernetes client is run within a Docker container, on Origin.)*
 *
 * @constructor
 */
function Client (conf) {
  this.conf = conf || {}

  // Attempt to fill in defaults if/when necessary
  if (conf) {
    if (!conf.apiBase)
      this.conf.apiBase = 'https://kubernetes.default.svc.cluster.local/api/v1'

    if (!conf.apiToken)
      this.conf.apiToken = fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf-8')
  }
}

Client.prototype.getServiceAccounts = function (done) {
  helpers.makeApiRequest(this.conf, {
    path: '/serviceaccounts'
  }, (err, res) => {
    if (err)
      done(err)
    else
      done(null, JSON.parse(res.text).items)
  })
}

module.exports = Client
