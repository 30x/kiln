'use strict'
const restify = require('restify');
const fs = require('fs');


const server = restify.createServer({
  name: 'myapp',
  version: '1.0.0'
});

/**
 * configure the server options
 */
server.use(restify.acceptParser(server.acceptable));
/**
 * Parse query params
 */
server.use( restify.queryParser() );
/**
 * only parse json body. TODO, check this for security
 */
server.use( restify.jsonBodyParser() );

 /**
  * Heartbeat for system liveliness
  * TODO integrate with origin/k8s call to ensure e2e communication
  */
server.get('/v1/heartbeat', function (req, res, next) {
  res.send({status:'ok'});
  return next();
});


/**
 * Get the logs
 */
server.get('/v1/logs/:mpname', function (req, res, next){
    const mpName = req.mpname;

    res.send(util.format("logs go here for %s", mpName));
    return next();
});

/**
 * The endpoint where the Zip file is posted to
 */
server.put('/v1/deploy/:name', function (req, res, next){


    //split for clarity
    const assetName = req.params.name;

    const someRandomeFileName = '/tmp/'+assetName + '.zip';

    const fileWriteStream = fs.createWriteStream(someRandomeFileName);

    const fileStream = req.pipe(fileWriteStream);

    fileStream.on('error', function(err){
        console.log(err)
    });

    //once we're done writing the stream, render a response
    fileStream.on('finish', function(){
        //TODO: Real work here

        //send back the endpoint the caller should hit for the deployed application
        res.send({endpoint:'http://endpointyouhit:8080'});

        return next();
    });


});


module.exports = {
    listen : function(port){
        server.listen(port, function () {
         console.log('%s listening at %s', server.name, server.url);
        });
    },

    close : function() {
      server.close();
    }
}
