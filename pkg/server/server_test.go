package server_test

import (
	"bytes"
	"net"
	"net/http"

	"encoding/json"
	"fmt"
	"time"

	"github.com/30x/shipyard/pkg/server"
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

		// Expect(err).Should(BeNil(), "Should not return an error creating a valid workspace")

		// //if could not find directory, it's a fail
		// Expect(workspace.SourceDirectory).Should(BeADirectory(), "Could not find directory "+workspace.SourceDirectory)

		// Expect(workspace.RootDirectory).Should(BeADirectory(), "Could not find directory "+workspace.RootDirectory)

		// Expect(workspace.SourceZipFile).ShouldNot(BeEmpty(), "SourceZipFile should be specified")

		// Expect(workspace.TargetTarName).ShouldNot(BeEmpty(), "TargetTarName should be specified")

		// Expect(workspace.DockerFile).Should(ContainSubstring(workspace.SourceDirectory), "Docker file should be in the source directory")

		// subString := strings.Replace(workspace.DockerFile, workspace.SourceDirectory, "", 1)

		// Expect(subString).Should(Equal("/Dockerfile"), "Dockerfile was not in the correct location")

	})
})
