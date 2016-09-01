package server_test

import (
	"bufio"
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

	"github.com/30x/kiln/pkg/server"
	"github.com/30x/kiln/pkg/kiln"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server Test", func() {

	BothContexts := func(testServer *server.Server, hostBase string, dockerRegistryURL string) {

		It("Get Imagespaces ", func() {

			httpResponse, imagespaces, err := getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			imagespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			response, body, err := newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//check build response
			sha, podspecURL := getBuildData(body)

			kiln.LogInfo.Printf("sha: %s\n podSpecUrl: %s\n", sha, podspecURL)

			Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			dockerURI := fmt.Sprintf("%s/%s/%s:%s", dockerRegistryURL, imagespace, application, revision)

			expectedURL := hostBase + fmt.Sprintf("/generatepodspec?imageURI=%s&publicPath=9000:/test-echo", dockerURI)

			Expect(podspecURL).Should(Equal(expectedURL), "Pod spec url should equal %s", expectedURL)

			fmt.Printf("Received PODSPEC URL of %s ", podspecURL)

			response, podSpec := getJSONContent(podspecURL)

			Expect(response.StatusCode).Should(Equal(200), "Get podspec at %s should equal 200", podspecURL)

			Expect(podSpec).ShouldNot(BeEmpty(), "Pod spec should have content")

			httpResponse, imagespaces, err = getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsNamespace(imagespaces, imagespace)

			//get the image
			response, imageSpec := getImage(hostBase, imagespace, application, revision)

			Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

			sha = imageSpec.ImageID

			Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

		})

		It("Get Imagespaces Trailing Slashes ", func() {

			httpResponse, imagespaces, err := getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			imagespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			response, body, err := newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo/")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//check build response
			sha, podspecURL := getBuildData(body)

			kiln.LogInfo.Printf("sha: %s\n podSpecUrl: %s\n", sha, podspecURL)

			Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			dockerURI := fmt.Sprintf("%s/%s/%s:%s", dockerRegistryURL, imagespace, application, revision)

			expectedURL := hostBase + fmt.Sprintf("/generatepodspec?imageURI=%s&publicPath=9000:/test-echo/", dockerURI)

			Expect(podspecURL).Should(Equal(expectedURL), "Pod spec url should equal %s", expectedURL)

			fmt.Printf("Received PODSPEC URL of %s ", podspecURL)

			response, podSpec := getJSONContent(podspecURL)

			Expect(response.StatusCode).Should(Equal(200), "Get podspec at %s should equal 200", podspecURL)

			Expect(podSpec).ShouldNot(BeEmpty(), "Pod spec should have content")

			httpResponse, imagespaces, err = getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsNamespace(imagespaces, imagespace)

		})

		It("Create Duplicate Application ", func() {
			//upload the first image
			imagespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			response, _, err := newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, applications, err := getApplications(hostBase, imagespace)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			assertContainsApplication(applications, application)

			//now try to post again, should get a 409

			response, _, err = newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(409), "409 should be returned")

		})

		It("Test Application Images", func() {
			//upload the first image
			imagespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			response, _, err := newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err := getImages(hostBase, imagespace, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 1).Should(BeTrue(), "Images should be >= than length 1")

			Expect(images[0].Created).ShouldNot(BeNil())

			Expect(images[0].ImageID).ShouldNot(BeNil())

			Expect(strings.Index(images[0].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			//now try to post again, with a new revision

			revision2 := "v1.1"

			response, _, err = newFileUploadRequest(hostBase, imagespace, application, revision2, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, images, err = getImages(hostBase, imagespace, application)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(images) >= 2).Should(BeTrue(), "Images should be of length >= 2")

			Expect(images[0].Created).ShouldNot(BeNil())

			Expect(images[0].ImageID).ShouldNot(BeNil())

			Expect(strings.Index(images[0].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			Expect(images[1].Created).ShouldNot(BeNil())

			Expect(images[1].ImageID).ShouldNot(BeNil())
			Expect(strings.Index(images[1].ImageID, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

		})

		It("No Cross Namepaces on GET", func() {
			//upload the first image
			imagespace1 := "test" + kiln.UUIDString()
			application1 := "application1"

			imagespace2 := "test" + kiln.UUIDString()
			application2 := "application2"

			revision := "v1.0"

			response, _, err := newFileUploadRequest(hostBase, imagespace1, application1, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//upload to imagespace 2 and ensure we can't see it
			response, _, err = newFileUploadRequest(hostBase, imagespace2, application2, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//now ensure it is created
			httpResponse, applications, err := getApplications(hostBase, imagespace1)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(applications)).Should(Equal(1), "Only 1 application should be present")

			assertContainsApplication(applications, application1)

			//ensure we only get the application 2

			httpResponse, applications, err = getApplications(hostBase, imagespace2)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(len(applications)).Should(Equal(1), "Only 1 application should be present")

			assertContainsApplication(applications, application2)

		})

		It("Podspec Storage", func() {

			namespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			content := `{"metadata":{"creationTimestamp":null,"labels":{"component":"test-application","routable":"true","runtime":"kiln"},"annotations":{"projectcalico.org/policy":"allow tcp from cidr 192.168.0.0/16; allow tcp from cidr 10.1.0.0/16","publicPaths":"9000:/example"}},"spec":{"volumes":null,"containers":[{"name":"test-application","image":"977777657611.dkr.ecr.us-west-2.amazonaws.com/test/example:0","ports":[{"containerPort":9000}],"env":[{"name":"PORT","value":"9000"},{"name":"PRIVATE_API_KEY","valueFrom":{"secretKeyRef":{"Name":"routing","key":"private-api-key"}}},{"name":"PUBLIC_API_KEY","valueFrom":{"secretKeyRef":{"Name":"routing","key":"public-api-key"}}}],"resources":{},"imagePullPolicy":"Always"}],"serviceAccountName":"","imagePullSecrets":[{"Name":"ecr-key"}]}}`

			httpResponse, body := putPodSpec(content, hostBase, namespace, application, revision)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(body).Should(Equal(""))

			//now get it

			httpResponse, body = getPodSpec(hostBase, namespace, application, revision)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(body).Should(Equal(content))

		})

		It("Podspec Storage", func() {

			namespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			content := `{"metadata":{"creationTimestamp":null,"labels":{"component":"test-application","routable":"true","runtime":"kiln"},"annotations":{"projectcalico.org/policy":"allow tcp from cidr 192.168.0.0/16; allow tcp from cidr 10.1.0.0/16","publicPaths":"9000:/example"}},"spec":{"volumes":null,"containers":[{"name":"test-application","image":"977777657611.dkr.ecr.us-west-2.amazonaws.com/test/example:0","ports":[{"containerPort":9000}],"env":[{"name":"PORT","value":"9000"},{"name":"PRIVATE_API_KEY","valueFrom":{"secretKeyRef":{"Name":"routing","key":"private-api-key"}}},{"name":"PUBLIC_API_KEY","valueFrom":{"secretKeyRef":{"Name":"routing","key":"public-api-key"}}}],"resources":{},"imagePullPolicy":"Always"}],"serviceAccountName":"","imagePullSecrets":[{"Name":"ecr-key"}]}}`

			httpResponse, body := putPodSpec(content, hostBase, namespace, application, revision)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(body).Should(Equal(""))

			//now get it

			httpResponse, body = getPodSpec(hostBase, namespace, application, revision)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			Expect(body).Should(Equal(content))

		})

	}

	ECROnly := func(testServer *server.Server, hostBase string, dockerRegistryURL string) {
		It("Delete Image ", func() {

			httpResponse, _, err := getImagespaces(hostBase)

			Expect(err).Should(BeNil(), "No error should be returned from the get. Error is %s", err)

			Expect(httpResponse.StatusCode).Should(Equal(200), "Response should be 200")

			//we explicity don't test since other images might be present in the docker registry

			//now create a new images

			imagespace := "test" + kiln.UUIDString()
			application := "application"
			revision := "v1.0"

			response, body, err := newFileUploadRequest(hostBase, imagespace, application, revision, "../../testresources/echo-test.zip", "9000:/test-echo")

			//read body to the end and discard results
			getBuildData(body)

			//do basic assertion before continuing
			Expect(err).Should(BeNil(), "Upload should be successfull")

			//now check the resposne code
			Expect(response.StatusCode).Should(Equal(201), "201 should be returned")

			//get the image
			response, imageSpec := getImage(hostBase, imagespace, application, revision)

			Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

			sha := imageSpec.ImageID

			Expect(strings.Index(sha, "sha256:")).Should(Equal(0), "Should start with sha256 signature")

			//now delete it
			response, imageSpec = deleteImage(hostBase, imagespace, application, revision)

			Expect(response.StatusCode).Should(Equal(200), "Get image should return 200")

			deleteSha := imageSpec.ImageID

			Expect(deleteSha).Should(Equal(sha), "Should start with sha256 signature")

			//should 404
			response, imageSpec = getImage(hostBase, imagespace, application, revision)

			Expect(response.StatusCode).Should(Equal(404), "Get image should return 404")

		})
	}

	Context("Local Docker", func() {
		//set up the provider

		dockerRegistryURL := "localhost:5000"

		os.Setenv("DOCKER_PROVIDER", "docker")
		os.Setenv("DOCKER_REGISTRY_URL", dockerRegistryURL)
		os.Setenv("POD_PROVIDER", "local")
		os.Setenv("LOCAL_DIR", "/tmp/podspecs")

		//Use our test provider for jwt tokens
		os.Setenv("JWTTOKENIMPL", "test")

		server, hostBase, err := doSetup(5280)

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		BothContexts(server, hostBase, dockerRegistryURL)
	})

	Context("ECR Docker", func() {

		dockerRegistryURL := "977777657611.dkr.ecr.us-east-1.amazonaws.com"

		//set up the provider
		os.Setenv("DOCKER_PROVIDER", "ecr")
		os.Setenv("DOCKER_REGISTRY_URL", dockerRegistryURL)
		os.Setenv("ECR_REGION", "us-east-1")
		os.Setenv("POD_PROVIDER", "local")
		os.Setenv("LOCAL_DIR", "/tmp/podspecs")

		//Use our test provider for jwt tokens
		os.Setenv("JWTTOKENIMPL", "test")

		server, hostBase, err := doSetup(5281)

		if err != nil {
			Fail(fmt.Sprintf("Could not start server %s", err))
		}

		BothContexts(server, hostBase, dockerRegistryURL)
		ECROnly(server, hostBase, dockerRegistryURL)
	})

})

//create a new instance of the server based on the env vars and the image creator.  Return them to be tested
func doSetup(port int) (*server.Server, string, error) {
	imageCreator, err := kiln.NewImageCreatorFromEnv()

	if err != nil {
		return nil, "", err
	}

	podSpecProvider, err := kiln.NewPodSpecIoFromEnv()

	if err != nil {
		return nil, "", err
	}

	baseHost := fmt.Sprintf("http://localhost:%d", port)

	testServer := server.NewServer(imageCreator, podSpecProvider)

	//start server in the background
	go func() {
		//start  the server and produce it to the start channel
		testServer.Start(port, 10*time.Second)
	}()

	//wait for it to start

	hostBase := fmt.Sprintf("%s/imagespaces", baseHost)

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
func assertContainsNamespace(imagespaces []*server.Imagespace, expectedImagespace string) {

	for _, imagespace := range imagespaces {
		if imagespace.Name == expectedImagespace {
			return
		}
	}

	Fail(fmt.Sprintf("Could not find imagespace %s in list returned", expectedImagespace))
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
func getImagespaces(hostBase string) (*http.Response, []*server.Imagespace, error) {

	url := fmt.Sprintf("%s/", hostBase)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return nil, nil, err
	}

	repositories := []*server.Imagespace{}

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, err
	}

	json.Unmarshal(bytes, &repositories)

	return response, repositories, err

}

//getApplications get the applications
func getApplications(hostBase string, imagespace string) (*http.Response, []*server.Application, error) {
	url := getApplicationsURL(hostBase, imagespace)
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
func getImages(hostBase string, imagespace string, application string) (*http.Response, []*server.Image, error) {

	url := getImagesURL(hostBase, imagespace, application)

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

func getPodSpec(hostBase string, namespace string, application string, revision string) (*http.Response, string) {
	url := getPodSpecURL(hostBase, namespace, application, revision)

	return getJSONContent(url)
}

func getImage(hostBase string, namespace string, application string, revision string) (*http.Response, *server.Image) {
	url := getImageURL(hostBase, namespace, application, revision)

	response, body := getJSONContent(url)

	image := &server.Image{}

	kiln.LogInfo.Printf("Response is %s", body)

	json.Unmarshal([]byte(body), image)

	return response, image

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
func newFileUploadRequest(hostBase string, imagespace string, application string, revision string, path string, publicPath string) (*http.Response, *string, error) {
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
	writer.WriteField("revision", revision)
	writer.WriteField("publicPath", publicPath)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	if err != nil {
		return nil, nil, err
	}

	uri := fmt.Sprintf("%s/%s/images", hostBase, imagespace)

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


//putPodSpec put the pod spec to the provided endpoint
func putPodSpec(podSpec string, hostBase string, namespace string, application string, revision string) (*http.Response, string) {
	url := getPodSpecURL(hostBase, namespace, application, revision)
	req, _ := http.NewRequest("PUT", url, strings.NewReader(podSpec))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, _ := client.Do(req)

	bytes, _ := ioutil.ReadAll(response.Body)

	return response, string(bytes)

}

//getPodSpecURL get the podspec url for GET and PUT operations
func getPodSpecURL(hostBase string, namespace string, application string, revision string) string {
	return fmt.Sprintf("%s/podspec/%s", getApplicationURL(hostBase, namespace, application), revision)
}

//getApplicationsURL get the appplicationsUrl
func getApplicationsURL(hostBase string, imagespace string) string {

	applicationsURL := fmt.Sprintf("%s/%s/images", hostBase, imagespace)

	// kiln.LogInfo.Printf("Creating URL %s", applicationsURL)

	return applicationsURL
}

func getApplicationURL(hostBase string, imagespace string, application string) string {
	return fmt.Sprintf("%s/%s", getApplicationsURL(hostBase, imagespace), application)
}

func getImagesURL(hostBase string, imagespace string, application string) string {
	return getApplicationURL(hostBase, imagespace, application)
}

//get the URL for the image
func getImageURL(hostBase string, imagespace string, application string, revision string) string {
	return fmt.Sprintf("%s/version/%s", getApplicationURL(hostBase, imagespace, application), revision)
}

func getBuildData(buildResponseBody *string) (imageSha string, podTemplateSpecURI string) {

	scanner := bufio.NewScanner(strings.NewReader(*buildResponseBody))

	scanner.Split(bufio.ScanLines)

	var line string

	for scanner.Scan() {

		line = scanner.Text()

		fmt.Print(line)

		if strings.HasPrefix(line, "ID:") {
			imageSha = strings.Replace(line, "ID:", "", -1)
			imageSha = strings.TrimSpace(imageSha)
		}

		if strings.HasPrefix(line, "PodTemplateSpec:") {
			podTemplateSpecURI = strings.Replace(line, "PodTemplateSpec:", "", -1)
			podTemplateSpecURI = strings.TrimSpace(podTemplateSpecURI)
		}

	}

	return imageSha, podTemplateSpecURI

}
