package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-swagger/go-swagger/errors"
	httpkit "github.com/go-swagger/go-swagger/httpkit"
	middleware "github.com/go-swagger/go-swagger/httpkit/middleware"

	"github.com/30x/shipyard/server/restapi/operations"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.ApisForBuildingDockerImagesAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ApisForBuildingDockerImagesAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.JSONConsumer = httpkit.JSONConsumer()

	api.JSONProducer = httpkit.JSONProducer()

	api.CreateApplicationHandler = operations.CreateApplicationHandlerFunc(func(params operations.CreateApplicationParams) middleware.Responder {
		return middleware.NotImplemented("operation .CreateApplication has not yet been implemented")
	})
	api.GetAllApplicationsHandler = operations.GetAllApplicationsHandlerFunc(func(params operations.GetAllApplicationsParams) middleware.Responder {
		return middleware.NotImplemented("operation .GetAllApplications has not yet been implemented")
	})
	api.GetImageHandler = operations.GetImageHandlerFunc(func(params operations.GetImageParams) middleware.Responder {
		return middleware.NotImplemented("operation .GetImage has not yet been implemented")
	})
	api.GetImageRevisionsHandler = operations.GetImageRevisionsHandlerFunc(func(params operations.GetImageRevisionsParams) middleware.Responder {
		return middleware.NotImplemented("operation .GetImageRevisions has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
