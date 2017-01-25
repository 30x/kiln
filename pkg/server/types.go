package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

var regex = regexp.MustCompile(`\d+:\/[\w+(\/)?]?`)

// DefaultRuntime the default version of alpine-node image
const DefaultRuntime = "node:4"

//Image represents an image struct
type Image struct {
	Created  *time.Time `json:"created,omitempty"`
	Revision []string   `json:"revision,omitempty"`
	ImageID  string     `json:"imageId,omitempty"`
}

//Organization represents an image struct
type Organization NamedObject

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
	Organization string
	Application  string   `schema:"name"`
	Revision     string   `schema:"revision"`
	EnvVars      []string `schema:"envVar"`
	Runtime      string   `schema:"runtime"`
}

//Validate validate the application input is correct
func (createImage *CreateImage) Validate() *Validation {
	errors := &Validation{
		messages: make(map[string]string),
	}

	if createImage.Organization == "" {
		errors.Add("Organization", "Organization must be specified")
	}

	if createImage.Application == "" {
		errors.Add("Application", "Application must be specified")
	}

	if createImage.Revision == "" {
		errors.Add("Revision", "Please enter a valid revision")
	}

	if createImage.Runtime == "" {
		createImage.Runtime = DefaultRuntime // default to node:4
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
