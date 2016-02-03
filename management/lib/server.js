'use strict'
const restify = require('restify');
 

const server = restify.createServer({
  name: 'myapp',
  version: '1.0.0'
});
server.use(restify.acceptParser(server.acceptable));
server.use(restify.queryParser());
server.use(restify.bodyParser());
 
server.get('/v1/management/heartbeat', function (req, res, next) {
  res.send({status:'ok'});
  return next();
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