'use strict'

const request = require('superagent')

// Disable SSL rejection due to Superagent bug
process.env.NODE_TLS_REJECT_UNAUTHORIZED = 0

/**
 * Makes an API request.
 *
 * @param {object} conf - The API client configuration
 * @param {options} options - The API request options
 * @param {function} done - The error-first callback
 */
module.exports.makeApiRequest = function (conf, options, done) {
  var verb = options.method || 'GET';
  var req;

  if (verb === 'DELETE')
    verb = 'DEL'

  verb = verb.toLowerCase()

  req = request[verb](conf.apiBase + options.path)

  if (options.query)
    req = req.query(options.query)

  if (options.headers) {
    Object.keys(options.headers).forEach((headerName) => {
      req = req.set(headerName, options.headers[headerName])
    })
  }

  if (options.body)
    req = req.send(options.body)

  // Set the authorization
  req = req.set('Authorization', 'Bearer ' + conf.apiToken)

  req.end((err, res) => {
    if (err || req.error)
      done(err || req.error)
    else
      done(null, res)
  })
}
