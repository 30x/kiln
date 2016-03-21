package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-swagger/go-swagger/httpkit"

	"github.com/30x/shipyard/server/models"
)

/*GetImageOK The request was for a valid repository and application and revision

swagger:response getImageOK
*/
type GetImageOK struct {

	// In: body
	Payload *models.Image `json:"body,omitempty"`
}

// NewGetImageOK creates GetImageOK with default headers values
func NewGetImageOK() *GetImageOK {
	return &GetImageOK{}
}

// WithPayload adds the payload to the get image o k response
func (o *GetImageOK) WithPayload(payload *models.Image) *GetImageOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get image o k response
func (o *GetImageOK) SetPayload(payload *models.Image) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetImageOK) WriteResponse(rw http.ResponseWriter, producer httpkit.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		if err := producer.Produce(rw, o.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetImageNotFound The request was for a repository or application that does not exist

swagger:response getImageNotFound
*/
type GetImageNotFound struct {

	// In: body
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetImageNotFound creates GetImageNotFound with default headers values
func NewGetImageNotFound() *GetImageNotFound {
	return &GetImageNotFound{}
}

// WithPayload adds the payload to the get image not found response
func (o *GetImageNotFound) WithPayload(payload *models.Error) *GetImageNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get image not found response
func (o *GetImageNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetImageNotFound) WriteResponse(rw http.ResponseWriter, producer httpkit.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		if err := producer.Produce(rw, o.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
