package server_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"
	"fmt"
	"time"

	"github.com/30x/shipyard/pkg/server"
	"github.com/30x/shipyard/pkg/shipyard"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server Test", func() {

	var testServer *server.Server
	var hostBase string

	//set up the server and client
	var _ = BeforeSuite(func() {
		port := 5280
		testServer = server.NewServer()

		//start server in the background
		go func() {
			//start  the server and produce it to the start channel
			err := testServer.Start(port)
			Expect(err).Should(BeNil(), "Could not start server on port %d, error is %s", port, err)
		}()

		//wait for it to start

		hostBase = fmt.Sprintf("http://localhost:%d/beeswax/images/api/v1", port)

		started := false

		//wait for host to start for 10 seconds
		for i := 0; i < 20; i++ {

			host := fmt.Sprintf("localhost:%d", port)

			conn, err := net.Dial("tcp", host)

			//done waiting, continue
			if err == nil {
				conn.Close()
				started = true
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		Expect(started).Should(BeTrue(), "Server should have started")

	})

	It("Get Namespaces ", func() {

		httpResponse, namespaces, err := getNamespaces(hostBase)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(httpResponse.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

		Expect(len(namespaces)).Should(Equal(0), "no namespaces should be present")

		//now create a new images

		namespace := "test" + shipyard.UUIDString()
		application := "application"
		revision := "v1.0"

		response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

		//do basic assertion before continuing
		Expect(err).Should(BeNil(), "Upload should be successfull")

		//now check the resposne code
		Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

		httpResponse, namespaces, err = getNamespaces(hostBase)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(httpResponse.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

		assertContainsNamespace(namespaces, namespace)

	})

	It("Create Duplicate Application ", func() {
		//upload the first image
		namespace := "test" + shipyard.UUIDString()
		application := "application"
		revision := "v1.0"

		response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

		//do basic assertion before continuing
		Expect(err).Should(BeNil(), "Upload should be successfull")

		//now check the resposne code
		Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

		//now ensure it is created
		httpResponse, applications, err := getApplications(hostBase, namespace)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(httpResponse.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

		assertContainsApplication(applications, application)

		//now try to post again, should get a 409

		response, err = newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

		//do basic assertion before continuing
		Expect(err).Should(BeNil(), "Upload should be successfull")

		//now check the resposne code
		Expect(response.StatusCode).Should(Equal(409), "409 should be returned")

	})

	It("Test Application Images", func() {
		//upload the first image
		namespace := "test" + shipyard.UUIDString()
		application := "application"
		revision := "v1.0"

		response, err := newFileUploadRequest(hostBase, namespace, application, revision, "../../testresources/echo-test.zip")

		//do basic assertion before continuing
		Expect(err).Should(BeNil(), "Upload should be successfull")

		//now check the resposne code
		Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

		//now ensure it is created
		httpResponse, images, err := getImages(hostBase, namespace, application)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(httpResponse.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

		Expect(len(images)).Should(Equal(1), "Images should be of length 1")

		Expect(images[0].Created).ShouldNot(BeNil())

		Expect(images[0].Size > 0).Should(BeTrue())

		Expect(images[0].ImageID).ShouldNot(BeNil())

		//now try to post again, with a new revision

		revision2 := "v1.1"

		response, err = newFileUploadRequest(hostBase, namespace, application, revision2, "../../testresources/echo-test.zip")

		//do basic assertion before continuing
		Expect(err).Should(BeNil(), "Upload should be successfull")

		//now check the resposne code
		Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

		//now ensure it is created
		httpResponse, images, err = getImages(hostBase, namespace, application)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(httpResponse.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

		Expect(len(images)).Should(Equal(2), "Images should be of length 2")

		Expect(images[0].Created).ShouldNot(BeNil())

		Expect(images[0].Size > 0).Should(BeTrue())

		Expect(images[0].ImageID).ShouldNot(BeNil())

		Expect(images[1].Created).ShouldNot(BeNil())

		Expect(images[1].Size > 0).Should(BeTrue())

		Expect(images[1].ImageID).ShouldNot(BeNil())

	})

})

//assertContainsNamespace check if namespace array has the expected namespace in it
func assertContainsNamespace(namespaces []*server.Namespace, expectedNamespace string) {

	for _, namespace := range namespaces {
		if namespace.Name == expectedNamespace {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find namespace %s in list returned", expectedNamespace))
}

func assertContainsApplication(applications []*server.Application, expectedApplication string) {

	for _, application := range applications {
		if application.Name == expectedApplication {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find application %s in list returned", expectedApplication))
}

//getNamespaces perform a get request on namespaces
func getNamespaces(hostBase string) (*http.Response, []*server.Namespace, error) {
	responseBuffer := &bytes.Buffer{}

	url := fmt.Sprintf("%s/namespaces", hostBase)

	req, err := http.NewRequest("GET", url, responseBuffer)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	repositories := []*server.Namespace{}

	json.Unmarshal(responseBuffer.Bytes(), &repositories)

	return response, repositories, err

}

//getApplications get the applications
func getApplications(hostBase string, namespace string) (*http.Response, []*server.Application, error) {
	responseBuffer := &bytes.Buffer{}

	url := getApplicationsURL(hostBase, namespace)
	req, err := http.NewRequest("GET", url, responseBuffer)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	repositories := []*server.Application{}

	json.Unmarshal(responseBuffer.Bytes(), &repositories)

	return response, repositories, err

}

//getImages get the images from the response
func getImages(hostBase string, namespace string, application string) (*http.Response, []*server.Image, error) {
	responseBuffer := &bytes.Buffer{}

	url := fmt.Sprintf("%s/images", getApplicationsURL(hostBase, namespace))
	req, err := http.NewRequest("GET", url, responseBuffer)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	repositories := []*server.Image{}

	json.Unmarshal(responseBuffer.Bytes(), &repositories)

	return response, repositories, err

}

//newfileUploadRequest upload a file form request
func newFileUploadRequest(hostBase string, namespace string, application string, revision string, path string) (*http.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))

	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	if err != nil {
		return nil, err
	}

	writer.WriteField("application", application)
	writer.WriteField("revision", revision)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	if err != nil {
		return nil, err
	}

	uri := getApplicationsURL(hostBase, namespace)

	request, err := http.NewRequest("POST", uri, body)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	return client.Do(request)
}

//getApplicationsURL get the appplicationsUrl
func getApplicationsURL(hostBase string, namespace string) string {

	applicationsURL := fmt.Sprintf("%s/namespaces/%s/applications", hostBase, namespace)

	shipyard.LogInfo.Printf("Getting URL %s", applicationsURL)

	return applicationsURL
}
