package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

var regex = regexp.MustCompile(`\d+:\/[\w+(\/)?]?`)
// DefaultNodeVersion the default version of alpine-node image
const DefaultNodeVersion = "4"

//Image represents an image struct
type Image struct {
	Created time.Time `json:"created"`
	ImageID string    `json:"imageId"`
}

//Imagespace represents an image struct
type Imagespace NamedObject

//Application represents an image struct
type Application NamedObject

//NamedObject An object that just contains name and links
type NamedObject struct {
	Name string `json:"name"`
}

//Link a link that represents a struct
type Link struct {
	Description string `json:"description"`
	Href        string `json:"href"`
}

//CreateImage the structure for creating an appliction via form
type CreateImage struct {
	Imagespace   string
	Application string `schema:"name"`
	Revision    string `schema:"revision"`
	PublicPath  string `schema:"publicPath"`
	EnvVars     []string `schema:"envVar"`
	NodeVersion string `schema:"nodeVersion"`
}

//Validate validate the application input is correct
func (createImage *CreateImage) Validate() *Validation {
	errors := &Validation{
		messages: make(map[string]string),
	}

	if createImage.Imagespace == "" {
		errors.Add("Imagespace", "Imagespace must be specified")
	}

	if createImage.Application == "" {
		errors.Add("Application", "Application must be specified")
	}

	if createImage.Revision == "" {
		errors.Add("Revision", "Please enter a valid revision")
	}

	if createImage.PublicPath == "" {
		errors.Add("PublicPath", "Please enter a valid publicPath.  None was specified.")
	}

	if !regex.Match([]byte(createImage.PublicPath)) {
		errors.Add("PublicPath", "Public path must match the format of [PORT]:/[URL SEGMENT/]?")
	}

	if createImage.NodeVersion == "" {
		createImage.NodeVersion = DefaultNodeVersion // default to alpine-node:4
	}

	return errors
}

//Error should be rendered when an error occurs
type Error struct {
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}

//Validation struct
type Validation struct {
	messages map[string]string
}

//HasErrors returns true if there are validation errors
func (validationFailure *Validation) HasErrors() bool {
	return len(validationFailure.messages) != 0
}

//Add add a field and an error
func (validationFailure *Validation) Add(field string, message string) {
	validationFailure.messages[field] = message
}

//WriteResponse write a response with validation errors
func (validationFailure *Validation) WriteResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)

	errors := []Error{}

	for _, message := range validationFailure.messages {

		errorObj := Error{
			Message: message,
		}

		errors = append(errors, errorObj)
	}

	json.NewEncoder(w).Encode(errors)
}
