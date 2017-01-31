package server_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/30x/kiln/pkg/kiln"
	"github.com/30x/kiln/pkg/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server Test", func() {

	BothContexts := func(testServer *server.Server, hostBase string, dockerRegistryURL string) {

		It("Get Organizations ", func() {

			httpResponse, organizations, err := getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			organization := "test" + kiln.UUIDString()
			application := "application"
			revision := "1"

			response, _, err := newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			httpResponse, organizations, err = getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsNamespace(organizations, organization)

			//get the image
			response, imageSpec := getImage(hostBase, organization, application, revision)

			Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

			sha := imageSpec.ImageID

			Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

		})

		It("Get Organizations Trailing Slashes ", func() {

			httpResponse, organizations, err := getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			organization := "test" + kiln.UUIDString()
			application := "application"

			response, _, err := newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			httpResponse, organizations, err = getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsNamespace(organizations, organization)

		})

		It("Test Application Images", func() {
			//upload the first image
			organization := "test" + kiln.UUIDString()
			application := "application"

			response, _, err := newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err := getImages(hostBase, organization, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 1).Should(BeTrue(), "Images should be >= than length 1")

			Expect(images[0].Revision).ShouldNot(BeNil())

			//now try to post again, with a new revision

			response, _, err = newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err = getImages(hostBase, organization, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 2).Should(BeTrue(), "Images should be of length >= 2")

			Expect(images[0].Revision).ShouldNot(BeNil())

			Expect(images[1].Revision).ShouldNot(BeNil())

		})

		It("No Cross Namepaces on GET", func() {
			//upload the first image
			organization1 := "test" + kiln.UUIDString()
			application1 := "application1"

			organization2 := "test" + kiln.UUIDString()
			application2 := "application2"

			response, _, err := newFileUploadRequest(hostBase, organization1, application1, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//upload to organization 2 and ensure we can't see it
			response, _, err = newFileUploadRequest(hostBase, organization2, application2, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, applications, err := getApplications(hostBase, organization1)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(applications)).Should(Equal(1), "Only 1 application should be present")

			assertContainsApplication(applications, application1)

			//ensure we only get the application 2

			httpResponse, applications, err = getApplications(hostBase, organization2)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(applications)).Should(Equal(1), "Only 1 application should be present")

			assertContainsApplication(applications, application2)

		})

	}

	// ECROnly := func(testServer *server.Server, hostBase string, dockerRegistryURL string) {
	// 	It("Delete Image ", func() {

	// 		httpResponse, _, err := getImagespaces(hostBase)

	// 		Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

	// 		Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

	// 		//we explicity don't test since other images might be present in the docker registry

	// 		//now create a new images

	// 		organization := "test" + kiln.UUIDString()
	// 		application := "application"
	// 		revision := "1"

	// 		response, _, err := newFileUploadRequest(hostBase, organization, application, revision, "../../testresources/echo-test.zip")

	// 		//do basic assertion before continuing
	// 		Expect(err).Should(BeNil(), "Upload should be successfull")

	// 		//now check the resposne code
	// 		Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

	// 		//get the image
	// 		response, imageSpec := getImage(hostBase, organization, application, revision)

	// 		Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

	// 		sha := imageSpec.ImageID

	// 		Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

	// 		//now delete it
	// 		response, imageSpec = deleteImage(hostBase, organization, application, revision)

	// 		Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

	// 		deleteSha := imageSpec.ImageID

	// 		Expect(deleteSha).Should(Equal(sha), "Should start with sha256 signature")

	// 		//should 404
	// 		response, imageSpec = getImage(hostBase, organization, application, revision)

	// 		Expect(response.StatusCode).Should(Equal(404), "Get image should return 404")

	// 	})
	// }

	ClusterDependent := func(testServer *server.Server, hostBase string, dockerRegistryURL string) {
		It("Delete Application ", func() {
			//upload the first image
			organization := "test-app-deletion-org"
			application := "deletemeplease"

			response, _, err := newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//upload the second image
			response, _, err = newFileUploadRequest(hostBase, organization, application, "../../testresources/echo-test.zip")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			// delete all images of application
			response = deleteApplication(hostBase, organization, application)

			Expect(response.StatusCode).Should(Equal(200), "200 should be returned")
		})
	}

	Context("Local Docker", func() {
		//set up the provider

		dockerRegistryURL := "localhost:5000"

		os.Setenv("DOCKER_PROVIDER", "docker")
		os.Setenv("DOCKER_REGISTRY_URL", dockerRegistryURL)

		//Use our test provider for jwt tokens
		os.Setenv("JWTTOKENIMPL", "test")

		server, hostBase, err := doSetup(5280, &kiln.ClusterConfig{})

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		BothContexts(server, hostBase, dockerRegistryURL)
	})

	// Context("ECR Docker", func() {

	// 	dockerRegistryURL := "977777657611.dkr.ecr.us-east-1.amazonaws.com"

	// 	//set up the provider
	// 	os.Setenv("DOCKER_PROVIDER", "ecr")
	// 	os.Setenv("DOCKER_REGISTRY_URL", dockerRegistryURL)
	// 	os.Setenv("ECR_REGION", "us-east-1")

	// 	//Use our test provider for jwt tokens
	// 	os.Setenv("JWTTOKENIMPL", "test")

	// 	server, hostBase, err := doSetup(5281)

	// 	if err != nil {
	// 		Fail(fmt.Sprintf("Could not start server %s", err))
	// 	}

	// 	BothContexts(server, hostBase, dockerRegistryURL)
	// 	ECROnly(server, hostBase, dockerRegistryURL)
	// })

	Context("With Local Cluster Config", func() {
		//set up the provider

		dockerRegistryURL := "localhost:5000"

		os.Setenv("DOCKER_PROVIDER", "docker")
		os.Setenv("DOCKER_REGISTRY_URL", dockerRegistryURL)

		os.Setenv("ORG_LABEL", "org")
		os.Setenv("APP_NAME_LABEL", "appName")

		//Use our test provider for jwt tokens
		os.Setenv("JWTTOKENIMPL", "test")

		clusterConfig, err := kiln.NewLocalClusterConfig()
		if err != nil {
			Fail(fmt.Sprintf("Could not get cluster config %s", err))
		}

		server, hostBase, err := doSetup(5282, clusterConfig)

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		ClusterDependent(server, hostBase, dockerRegistryURL)
	})

})

//create a new instance of the server based on the env vars and the image creator.  Return them to be tested
func doSetup(port int, clusterConfig *kiln.ClusterConfig) (*server.Server, string, error) {
	imageCreator, err := kiln.NewImageCreatorFromEnv()

	if err != nil {
		return nil, "", err
	}

	baseHost := fmt.Sprintf("http://localhost:%d", port)

	testServer := server.NewServer(imageCreator, clusterConfig)

	//start server in the background
	go func() {
		//start  the server and produce it to the start channel
		testServer.Start(port, 10*time.Second)
	}()

	//wait for it to start

	hostBase := fmt.Sprintf("%s/organizations", baseHost)

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

	if !started {
		return nil, "", errors.New("Server did not start")
	}

	return testServer, hostBase, nil
}

//assertContainsNamespace check if namespace array has the expected namespace in it
func assertContainsNamespace(organizations []*server.Organization, expectedOrganization string) {

	for _, organziation := range organizations {
		if organziation.Name == expectedOrganization {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find organziation %s in list returned", expectedOrganization))
}

func assertContainsApplication(applications []*server.Application, expectedApplication string) {

	for _, application := range applications {
		if application.Name == expectedApplication {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find application %s in list returned", expectedApplication))
}

//getImagespaces perform a get request on namespaces
func getImagespaces(hostBase string) (*http.Response, []*server.Organization, error) {

	url := fmt.Sprintf("%s/", hostBase)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return nil, nil, err
	}

	repositories := []*server.Organization{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	json.Unmarshal(bytes, &repositories)

	return response, repositories, err

}

//getApplications get the applications
func getApplications(hostBase string, organization string) (*http.Response, []*server.Application, error) {
	url := getApplicationsURL(hostBase, organization)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	client := &http.Client{}
	response, err := client.Do(req)

	repositories := []*server.Application{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	json.Unmarshal(bytes, &repositories)

	return response, repositories, err

}

//getImages get the images from the response
func getImages(hostBase string, organization string, application string) (*http.Response, []*server.Image, error) {

	url := getImagesURL(hostBase, organization, application)

	kiln.LogInfo.Printf("Invoking get at URL %s", url)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	client := &http.Client{}
	response, err := client.Do(req)

	images := []*server.Image{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	body := string(bytes)

	kiln.LogInfo.Printf("Response is %s", body)

	json.Unmarshal(bytes, &images)

	return response, images, err

}

func getJSONContent(url string) (*http.Response, string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	client := &http.Client{}
	response, _ := client.Do(req)

	bytes, _ := ioutil.ReadAll(response.Body)

	return response, string(bytes)

}

func getImage(hostBase string, namespace string, application string, revision string) (*http.Response, *server.Image) {
	url := getImageURL(hostBase, namespace, application, revision)

	response, body := getJSONContent(url)

	image := &server.Image{}

	kiln.LogInfo.Printf("Response is %s", body)

	json.Unmarshal([]byte(body), image)

	return response, image

}

func deleteApplication(hostBase string, namespace string, application string) *http.Response {
	url := getApplicationURL(hostBase, namespace, application)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	response, _ := client.Do(req)

	bytes, _ := ioutil.ReadAll(response.Body)

	body := string(bytes)

	kiln.LogInfo.Printf("Response is %s", body)

	return response
}

func deleteImage(hostBase string, namespace string, application string, revision string) (*http.Response, *server.Image) {
	url := getImageURL(hostBase, namespace, application, revision)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	response, _ := client.Do(req)

	bytes, _ := ioutil.ReadAll(response.Body)

	image := &server.Image{}

	kiln.LogInfo.Printf("Response is %s", bytes)

	json.Unmarshal(bytes, image)

	return response, image

}

//newfileUploadRequest upload a file form request. Returns the response, the fully read body as a string, and an error
func newFileUploadRequest(hostBase string, organization string, application string, path string) (*http.Response, *string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))

	if err != nil {
		return nil, nil, err
	}
	_, err = io.Copy(part, file)

	if err != nil {
		return nil, nil, err
	}

	writer.WriteField("name", application)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	if err != nil {
		return nil, nil, err
	}

	uri := fmt.Sprintf("%s/%s/apps", hostBase, organization)

	request, err := http.NewRequest("POST", uri, body)

	if err != nil {
		return nil, nil, err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	response, err := client.Do(request)

	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	bodyResponse := string(bodyBytes)

	return response, &bodyResponse, nil
}

//getApplicationsURL get the appplicationsUrl
func getApplicationsURL(hostBase string, organization string) string {

	applicationsURL := fmt.Sprintf("%s/%s/apps", hostBase, organization)

	// kiln.LogInfo.Printf("Creating URL %s", applicationsURL)

	return applicationsURL
}

func getApplicationURL(hostBase string, organization string, application string) string {
	return fmt.Sprintf("%s/%s", getApplicationsURL(hostBase, organization), application)
}

func getImagesURL(hostBase string, organization string, application string) string {
	return getApplicationURL(hostBase, organization, application)
}

//get the URL for the image
func getImageURL(hostBase string, organization string, application string, revision string) string {
	return fmt.Sprintf("%s/version/%s", getApplicationURL(hostBase, organization, application), revision)
}
