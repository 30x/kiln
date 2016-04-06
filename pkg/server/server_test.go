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
	var client *http.Client

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

		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
		}

	})

	It("Get Repositories ", func() {

		response := &bytes.Buffer{}

		req, err := http.NewRequest("GET", hostBase+"/namespaces", response)
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)

		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

		Expect(resp.ContentLength > 0).Should(BeTrue(), "Should have a response body")

		//unmarshall
		repositories := []*server.Namespace{}

		json.Unmarshal(response.Bytes(), &repositories)

		Expect(len(repositories)).Should(Equal(0), "no repositories should be present")

		//now create a new images

		namespace := "test" + shipyard.UUIDString()
		application := "application"
		revision := "v1.0"

		req, err = http.NewRequest("POST", getApplicationsUrl(hostBase, namespace), response)
		req.Header.Add("Content-Type", "application/json")

		//add the form parameters
		req.Form.Add("application", application)
		req.Form.Add("revision", revision)

		resp, err = client.Do(req)

	})
})

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return http.NewRequest("POST", uri, body)
}

//getApplicationsUrl get the appplicationsUrl
func getApplicationsUrl(hostBase string, namespace string) string {

	return fmt.Sprintf("%s/namespaces/%s/applications/", hostBase, namespace)
}
