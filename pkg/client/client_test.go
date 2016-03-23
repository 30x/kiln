package client_test

import (
	"os"
	"testing"

	"github.com/30x/shipyard/pkg/client/client"
	"github.com/30x/shipyard/pkg/client/client/operations"
	httptransport "github.com/go-swagger/go-swagger/httpkit/client"
	"github.com/go-swagger/go-swagger/strfmt"
)

// NewHTTPClient creates a new apis for building docker images HTTP client.

func TestCreateImage(t *testing.T) {

	transport := httptransport.New("localhost:5280", "/beeswax/images/api/v1", []string{"http"})

	formats := strfmt.Default

	client := client.New(transport, formats)

	createApplicationParams := operations.NewCreateApplicationParams()
	createApplicationParams.Repository = "Test"
	createApplicationParams.Application = "echo1"
	createApplicationParams.Revision = "v1.0"

	file, err := os.Open("../../testresources/echo-test.zip")

	if err != nil {
		t.Fatal("Could not open file for upload")
	}

	createApplicationParams.File = *file

	response, err := client.Operations.CreateApplication(createApplicationParams)

	if err != nil {
		t.Fatal("Could not create application", err)
	}

	image := response.Payload

	if image.ImageID == "" {
		t.Fatal("Didn't do anything")
	}

}
