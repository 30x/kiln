package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

//Image represents an image struct
type Image struct {
	Created time.Time `json:"created"`
	Size    int64     `json:"size"`
	ImageID string    `json:"imageId"`
	Links   []Link    `json:"_links"`
}

//Application represents an image struct
type Application struct {
	Name  string `json:"name"`
	Links []Link `json:"_links"`
}

//Applications An array of applications
type Applications []Application

//Link a link that represents a struct
type Link struct {
	Description string `json:"description"`
	Href        string `json:"href"`
}

//CreateApplication the structure for creating an appliction via form
type CreateApplication struct {
	Application string  `schema:"application"`
	Revision    string  `schema:"revision"`
	File        os.File `schema:"file"`
}

//Validate validate the application input is correct
func (createApplication *CreateApplication) Validate() *Validation {
	errors := &Validation{
		messages: make(map[string]string),
	}

	// re := regexp.MustCompile(".+@.+\\..+")
	// matched := re.Match([]byte(msg.Email))

	if createApplication.Application == "" {
		errors.Add("Application", "Application must be specified")
	}

	if createApplication.Revision == "" {
		errors.Add("Revision", "Please enter a valid revision")
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
