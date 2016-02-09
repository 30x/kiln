'use strict';

/**
 * Kubernetes client wrapper for shipyard
 * @param k8sHost
 * @constructor
 */
function Kubernetes(k8sHost) {

}

module.exports = Kubernetes;


/**
 * Create the replication controller and service, invoke the callback with the success or fail
 * @param org
 * @param env
 * @param app
 * @param revision
 * @param dockerImageName
 * @param cb A function of (err, rcName)
 */
Kubernetes.prototype.createReplicationController = function (org, env, app, revision, dockerImageName, cb) {

};


/**
 * Delete the replication controller
 * @param org
 * @param env
 * @param app
 * @param revision
 * @param cb  A function of (err) where err is empty on succesfull delete
 */
Kubernetes.prototype.deleteReplicationController = function (org, env, app, revision, cb) {

};

/**
 * Create a service that uses the specified selectors for PODS.  Note that revision is intentionally omitted
 * this allows the service to point to multiple versions as a new replication controller is created
 * @param org
 * @param env
 * @param app
 * @param cb A function of (err, serviceIp)
 */
Kubernetes.prototype.createService = function (org, env, app, cb) {

};


/**
 * Delete the replication controller
 * @param org
 * @param env
 * @param app
 * @param revision
 * @param cb A function of type (err, logs) invoked when logs are returned
 */
Kubernetes.prototype.getLogs = function (org, env, app, revision, cb) {

};
