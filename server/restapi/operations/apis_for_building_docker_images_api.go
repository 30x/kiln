package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	httpkit "github.com/go-swagger/go-swagger/httpkit"
	middleware "github.com/go-swagger/go-swagger/httpkit/middleware"
	spec "github.com/go-swagger/go-swagger/spec"
	strfmt "github.com/go-swagger/go-swagger/strfmt"
	"github.com/go-swagger/go-swagger/swag"
)

// NewApisForBuildingDockerImagesAPI creates a new ApisForBuildingDockerImages instance
func NewApisForBuildingDockerImagesAPI(spec *spec.Document) *ApisForBuildingDockerImagesAPI {
	o := &ApisForBuildingDockerImagesAPI{
		spec:            spec,
		handlers:        make(map[string]map[string]http.Handler),
		formats:         strfmt.Default,
		defaultConsumes: "application/json",
		defaultProduces: "application/json",
		ServerShutdown:  func() {},
	}

	return o
}

/*ApisForBuildingDockerImagesAPI the apis for building docker images API */
type ApisForBuildingDockerImagesAPI struct {
	spec            *spec.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	defaultConsumes string
	defaultProduces string
	// JSONConsumer registers a consumer for a "application/json" mime type
	JSONConsumer httpkit.Consumer

	// JSONProducer registers a producer for a "application/json" mime type
	JSONProducer httpkit.Producer

	// CreateApplicationHandler sets the operation handler for the create application operation
	CreateApplicationHandler CreateApplicationHandler
	// GetAllApplicationsHandler sets the operation handler for the get all applications operation
	GetAllApplicationsHandler GetAllApplicationsHandler
	// GetImageHandler sets the operation handler for the get image operation
	GetImageHandler GetImageHandler
	// GetImageRevisionsHandler sets the operation handler for the get image revisions operation
	GetImageRevisionsHandler GetImageRevisionsHandler

	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup
}

// SetDefaultProduces sets the default produces media type
func (o *ApisForBuildingDockerImagesAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *ApisForBuildingDockerImagesAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// DefaultProduces returns the default produces media type
func (o *ApisForBuildingDockerImagesAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *ApisForBuildingDockerImagesAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *ApisForBuildingDockerImagesAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *ApisForBuildingDockerImagesAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the ApisForBuildingDockerImagesAPI
func (o *ApisForBuildingDockerImagesAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.CreateApplicationHandler == nil {
		unregistered = append(unregistered, "CreateApplicationHandler")
	}

	if o.GetAllApplicationsHandler == nil {
		unregistered = append(unregistered, "GetAllApplicationsHandler")
	}

	if o.GetImageHandler == nil {
		unregistered = append(unregistered, "GetImageHandler")
	}

	if o.GetImageRevisionsHandler == nil {
		unregistered = append(unregistered, "GetImageRevisionsHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *ApisForBuildingDockerImagesAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *ApisForBuildingDockerImagesAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]httpkit.Authenticator {

	return nil

}

// ConsumersFor gets the consumers for the specified media types
func (o *ApisForBuildingDockerImagesAPI) ConsumersFor(mediaTypes []string) map[string]httpkit.Consumer {

	result := make(map[string]httpkit.Consumer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONConsumer

		}
	}
	return result

}

// ProducersFor gets the producers for the specified media types
func (o *ApisForBuildingDockerImagesAPI) ProducersFor(mediaTypes []string) map[string]httpkit.Producer {

	result := make(map[string]httpkit.Producer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONProducer

		}
	}
	return result

}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *ApisForBuildingDockerImagesAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

func (o *ApisForBuildingDockerImagesAPI) initHandlerCache() {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/{repository}/applications"] = NewCreateApplication(o.context, o.CreateApplicationHandler)

	if o.handlers["GET"] == nil {
		o.handlers[strings.ToUpper("GET")] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/{repository}/applications"] = NewGetAllApplications(o.context, o.GetAllApplicationsHandler)

	if o.handlers["GET"] == nil {
		o.handlers[strings.ToUpper("GET")] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/{repository}/apps/{application}/images/{revision}"] = NewGetImage(o.context, o.GetImageHandler)

	if o.handlers["GET"] == nil {
		o.handlers[strings.ToUpper("GET")] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/{repository}/apps/{application}/images"] = NewGetImageRevisions(o.context, o.GetImageRevisionsHandler)

}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *ApisForBuildingDockerImagesAPI) Serve(builder middleware.Builder) http.Handler {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}

	return o.context.APIHandler(builder)
}
