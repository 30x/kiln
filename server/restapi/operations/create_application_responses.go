package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-swagger/go-swagger/httpkit"

	"github.com/30x/shipyard/server/models"
)

/*CreateApplicationOK The request was for a valid repo, application, and image

swagger:response createApplicationOK
*/
type CreateApplicationOK struct {

	// In: body
	Payload *models.Image `json:"body,omitempty"`
}

// NewCreateApplicationOK creates CreateApplicationOK with default headers values
func NewCreateApplicationOK() *CreateApplicationOK {
	return &CreateApplicationOK{}
}

// WithPayload adds the payload to the create application o k response
func (o *CreateApplicationOK) WithPayload(payload *models.Image) *CreateApplicationOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create application o k response
func (o *CreateApplicationOK) SetPayload(payload *models.Image) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateApplicationOK) WriteResponse(rw http.ResponseWriter, producer httpkit.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		if err := producer.Produce(rw, o.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*CreateApplicationNotFound The repository does not exist

swagger:response createApplicationNotFound
*/
type CreateApplicationNotFound struct {

	// In: body
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateApplicationNotFound creates CreateApplicationNotFound with default headers values
func NewCreateApplicationNotFound() *CreateApplicationNotFound {
	return &CreateApplicationNotFound{}
}

// WithPayload adds the payload to the create application not found response
func (o *CreateApplicationNotFound) WithPayload(payload *models.Error) *CreateApplicationNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create application not found response
func (o *CreateApplicationNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateApplicationNotFound) WriteResponse(rw http.ResponseWriter, producer httpkit.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		if err := producer.Produce(rw, o.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*CreateApplicationConflict Application already exists

swagger:response createApplicationConflict
*/
type CreateApplicationConflict struct {

	// In: body
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateApplicationConflict creates CreateApplicationConflict with default headers values
func NewCreateApplicationConflict() *CreateApplicationConflict {
	return &CreateApplicationConflict{}
}

// WithPayload adds the payload to the create application conflict response
func (o *CreateApplicationConflict) WithPayload(payload *models.Error) *CreateApplicationConflict {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create application conflict response
func (o *CreateApplicationConflict) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateApplicationConflict) WriteResponse(rw http.ResponseWriter, producer httpkit.Producer) {

	rw.WriteHeader(409)
	if o.Payload != nil {
		if err := producer.Produce(rw, o.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
