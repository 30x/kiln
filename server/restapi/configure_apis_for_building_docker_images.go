package restapi

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	errors "github.com/go-swagger/go-swagger/errors"
	httpkit "github.com/go-swagger/go-swagger/httpkit"
	middleware "github.com/go-swagger/go-swagger/httpkit/middleware"
	strfmt "github.com/go-swagger/go-swagger/strfmt"

	"github.com/30x/shipyard/server/models"
	"github.com/30x/shipyard/server/restapi/operations"
	"github.com/30x/shipyard/shipyard"
)

// This file is safe to edit. Once it exists it will not be overwritten
var imageCreator shipyard.ImageCreator

func configureFlags(api *operations.ApisForBuildingDockerImagesAPI) {
	var error error
	imageCreator, error = shipyard.NewEcsImageCreator("977777657611.dkr.ecr.us-east-1.amazonaws.com", "us-east-1")

	//we should die here if we're unable to start
	if error != nil {
		shipyard.LogError.Fatalf("Unable to create image creator %s", error)
	}

}

func configureAPI(api *operations.ApisForBuildingDockerImagesAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.JSONConsumer = httpkit.JSONConsumer()

	api.JSONProducer = httpkit.JSONProducer()

	api.CreateApplicationHandler = operations.CreateApplicationHandlerFunc(func(params operations.CreateApplicationParams) middleware.Responder {

		workspace, err := shipyard.CreateNewWorkspace()

		if err != nil {
			message := fmt.Sprintf("Unable to create workspace, %s", err)
			return InternalError(message)
		}

		//remove workspace after the request completes
		defer workspace.Clean()

		//copy the file data to a zip file
		base64FileData := params.File

		byteData := []byte{}

		err = base64FileData.UnmarshalText(byteData)

		if err != nil {
			message := fmt.Sprintf("Unable to unmarshall base64 into bytes %s", err)
			return InternalError(message)
		}

		//get the zip file and write bytes to it
		err = workspace.WriteZipeFileData(byteData)

		if err != nil {
			message := fmt.Sprintf("Unable to write zip file %s", err)
			return InternalError(message)
		}

		dockerInfo := &shipyard.DockerInfo{
			RepoName:  params.Repository,
			ImageName: params.Application,
			Revision:  params.Revision,
		}

		dockerFile := &shipyard.DockerFile{
			ParentImage: "node:4.3.0-onbuild",
			DockerInfo:  dockerInfo,
		}

		err = workspace.CreateDockerFile(dockerFile)

		if err != nil {
			message := fmt.Sprintf("Could not create docker file %s ", err)
			return InternalError(message)
		}

		err = workspace.BuildTarFile()

		if err != nil {
			message := fmt.Sprintf("Could not create tar file %s", err)
			return InternalError(message)
		}

		dockerBuild := &shipyard.DockerBuild{
			DockerInfo: dockerInfo,
			TarFile:    workspace.TargetTarName,
		}

		//TODO make this a real writes
		logWriter := os.Stdout

		err = imageCreator.BuildImage(dockerBuild, logWriter)

		if err != nil {
			message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
			return InternalError(message)
		}

		response := operations.NewCreateApplicationOK()

		response.Payload = &models.Image{
			Created: strfmt.NewDateTime(),
			ImageID: dockerInfo.GetImageName(),
		}

		return response
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
