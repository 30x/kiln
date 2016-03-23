package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-swagger/go-swagger/client"
	"github.com/go-swagger/go-swagger/errors"

	strfmt "github.com/go-swagger/go-swagger/strfmt"
)

// NewGetImageParams creates a new GetImageParams object
// with the default values initialized.
func NewGetImageParams() *GetImageParams {
	var ()
	return &GetImageParams{}
}

/*GetImageParams contains all the parameters to send to the API endpoint
for the get image operation typically these are written to a http.Request
*/
type GetImageParams struct {

	/*Application
	  The Application name

	*/
	Application string
	/*Repository
	  The Docker repository name

	*/
	Repository string
	/*Revision
	  The revision of the application

	*/
	Revision string
}

// WithApplication adds the application to the get image params
func (o *GetImageParams) WithApplication(application string) *GetImageParams {
	o.Application = application
	return o
}

// WithRepository adds the repository to the get image params
func (o *GetImageParams) WithRepository(repository string) *GetImageParams {
	o.Repository = repository
	return o
}

// WithRevision adds the revision to the get image params
func (o *GetImageParams) WithRevision(revision string) *GetImageParams {
	o.Revision = revision
	return o
}

// WriteToRequest writes these params to a swagger request
func (o *GetImageParams) WriteToRequest(r client.Request, reg strfmt.Registry) error {

	var res []error

	// path param application
	if err := r.SetPathParam("application", o.Application); err != nil {
		return err
	}

	// path param repository
	if err := r.SetPathParam("repository", o.Repository); err != nil {
		return err
	}

	// path param revision
	if err := r.SetPathParam("revision", o.Revision); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
