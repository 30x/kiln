package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-swagger/go-swagger/client"
	"github.com/go-swagger/go-swagger/errors"

	strfmt "github.com/go-swagger/go-swagger/strfmt"
)

// NewCreateApplicationParams creates a new CreateApplicationParams object
// with the default values initialized.
func NewCreateApplicationParams() *CreateApplicationParams {
	var ()
	return &CreateApplicationParams{}
}

/*CreateApplicationParams contains all the parameters to send to the API endpoint
for the create application operation typically these are written to a http.Request
*/
type CreateApplicationParams struct {

	/*Application
	  The Application name

	*/
	Application string
	/*File
	  The file data as a multipart

	*/
	File strfmt.Base64
	/*Repository
	  The Docker repository name

	*/
	Repository string
	/*Revision
	  The Revision of the image

	*/
	Revision string
}

// WithApplication adds the application to the create application params
func (o *CreateApplicationParams) WithApplication(application string) *CreateApplicationParams {
	o.Application = application
	return o
}

// WithFile adds the file to the create application params
func (o *CreateApplicationParams) WithFile(file strfmt.Base64) *CreateApplicationParams {
	o.File = file
	return o
}

// WithRepository adds the repository to the create application params
func (o *CreateApplicationParams) WithRepository(repository string) *CreateApplicationParams {
	o.Repository = repository
	return o
}

// WithRevision adds the revision to the create application params
func (o *CreateApplicationParams) WithRevision(revision string) *CreateApplicationParams {
	o.Revision = revision
	return o
}

// WriteToRequest writes these params to a swagger request
func (o *CreateApplicationParams) WriteToRequest(r client.Request, reg strfmt.Registry) error {

	var res []error

	// form param application
	frApplication := o.Application
	fApplication := frApplication
	if fApplication != "" {
		if err := r.SetFormParam("application", fApplication); err != nil {
			return err
		}
	}

	// form param file
	frFile := o.File
	fFile := frFile.String()
	if fFile != "" {
		if err := r.SetFormParam("file", fFile); err != nil {
			return err
		}
	}

	// path param repository
	if err := r.SetPathParam("repository", o.Repository); err != nil {
		return err
	}

	// form param revision
	frRevision := o.Revision
	fRevision := frRevision
	if fRevision != "" {
		if err := r.SetFormParam("revision", fRevision); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
