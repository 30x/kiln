'use strict';

var app = require('connect')();
var fs = require('fs');
var http = require('http');
var https = require('https');

var apiHost = 'openshift.default.svc.cluster.local';
var apiPathBase = '/oapi/v1';
var apiScheme = 'https';
var apiToken = fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf-8');
var port = process.env.PORT || process.env.OPENSHIFT_NODEJS_PORT || 8080;

app.use(function (req, res, next) {
  var data = '';
  var cReq = https.request({
    hostname: apiHost,
    path: apiPathBase + '/projects',
    port: 443,
    method: 'GET',
    headers: {
      'Authorization': 'Bearer ' + apiToken
    },
    rejectUnauthorized: false
  }, function (cRes) {
    cRes.setEncoding('utf-8');

    cRes.on('data', function (chunk) {
      data += chunk;
    });
    cRes.on('end', function () {
      res.setHeader('content-type', 'application/json');
      res.end(data);
    });
  });

  cReq.on('error', function (err) {
    console.error(err.stack);

    next(err);
  });

  cReq.end();
});

http.createServer(app).listen(port, function () {
  console.log('API Token: ', apiToken);
  console.log();
  console.log('Environment Variables');
  Object.keys(process.env).forEach(function (key) {
    console.log('  ', key, ': ' + process.env[key]);
  });
  console.log();
  console.log('Server listening on ', port);
});