package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/30x/authsdk"
	"github.com/30x/kiln/pkg/kiln"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/tylerb/graceful"
)

//TODO make an env variable.  100 Meg max
const maxFileSize = 1024 * 1024 * 100

const basePath = "/organizations"

//Server struct to create an instance of hte server
type Server struct {
	router        http.Handler
	decoder       *schema.Decoder
	imageCreator  kiln.ImageCreator
	clusterConfig *kiln.ClusterConfig
}

//NewServer Create a new server using the provided Image creator.
func NewServer(imageCreator kiln.ImageCreator, clusterConfig *kiln.ClusterConfig) *Server {
	routes := mux.NewRouter()

	//allow the trailing slash
	//Note that when setting this to true, all URLS must end with a slash so we match paths both with and without the trailing slash
	// routes.StrictSlash(true)

	//now set up the decoder and return the server
	decoder := schema.NewDecoder()

	server := &Server{
		decoder:       decoder,
		imageCreator:  imageCreator,
		clusterConfig: clusterConfig,
	}

	//a bit hacky, but need the pointer to the server
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/organizations/{org}/apps/").HandlerFunc(server.postApplication)
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/organizations/{org}/apps").HandlerFunc(server.postApplication)
	routes.Methods("GET").Path("/organizations/").HandlerFunc(server.getOrganizations) // getOrganization
	routes.Methods("GET").Path("/organizations").HandlerFunc(server.getOrganizations)
	routes.Methods("GET").Path("/organizations/{org}/apps/").HandlerFunc(server.getApplications) // get all images by name in organization
	routes.Methods("GET").Path("/organizations/{org}/apps").HandlerFunc(server.getApplications)
	routes.Methods("GET").Path("/organizations/{org}/apps/{name}/").HandlerFunc(server.getImages) // get all revisions of an app in an organization
	routes.Methods("GET").Path("/organizations/{org}/apps/{name}").HandlerFunc(server.getImages)
	routes.Methods("GET").Path("/organizations/{org}/apps/{name}/version/{revision}/").HandlerFunc(server.getImage) // get image by revision
	routes.Methods("GET").Path("/organizations/{org}/apps/{name}/version/{revision}").HandlerFunc(server.getImage)
	routes.Methods("DELETE").Path("/organizations/{org}/apps/{name}/").HandlerFunc(server.deleteApplication) // delete all app images
	routes.Methods("DELETE").Path("/organizations/{org}/apps/{name}").HandlerFunc(server.deleteApplication)

	// dockerfile
	routes.Methods("GET").Path("/organizations/kiln/Dockerfile/").HandlerFunc(server.getDockerfile)
	routes.Methods("GET").Path("/organizations/kiln/Dockerfile").HandlerFunc(server.getDockerfile)

	//health check
	routes.Methods("GET").Path("/organizations/status/").HandlerFunc(server.status)
	routes.Methods("GET").Path("/organizations/status").HandlerFunc(server.status)
	routes.Methods("GET").Path("/organizations/kiln/status/").HandlerFunc(server.status)
	routes.Methods("GET").Path("/organizations/kiln/status").HandlerFunc(server.status)

	//now wrap everything with logging

	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, routes)

	server.router = loggedRouter

	return server

}

//Start start the http server with the port and the specified timeout, as well as the shutdown timer
func (server *Server) Start(port int, timeout time.Duration) {
	address := fmt.Sprintf(":%d", port)

	kiln.LogInfo.Printf("Starting server at address %s", address)

	srv := &graceful.Server{
		Timeout: timeout,
		Server:  &http.Server{Addr: address, Handler: server.router},
		Logger:  graceful.DefaultLogger(),
	}

	//set up our timer in a gofunc in order to shut down after a duration

	//start listening
	if err := srv.ListenAndServe(); err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			srv.Logger.Fatal(err)
		}
	}

}

// getDockerFile replies with the Dockerfile kiln uses to build images
func (server *Server) getDockerfile(w http.ResponseWriter, r *http.Request) {
	dockerInfo := &kiln.DockerInfo{
		RepoName:  "<organization>",
		ImageName: "<imageName>",
		Revision:  "<revision>",
		EnvVars:   []string{"var1=val1", "var2=val2"},
		BaseImage: "<baseImage>",
	}

	resp, err := kiln.GetExampleDockerfile(dockerInfo)
	if err != nil {
		message := fmt.Sprintf("Failed making example Dockerfile %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(resp.Bytes())
}

// TODO: Fix how org and other info is parsed
//postApplication and render a response
func (server *Server) postApplication(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	// file, handler, err := r.FormFile("file")
	file, _, err := r.FormFile("file")

	if err != nil {
		message := fmt.Sprintf("Unable to upload file %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//defer closing after request completes
	defer file.Close()

	createImage := new(CreateImage)

	// pull org name, from path vars
	vars := mux.Vars(r)
	createImage.Organization = vars["org"]

	// r.PostForm is a map of our POST form values without the file
	err = server.decoder.Decode(createImage, r.Form)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	validation := createImage.Validate()

	if validation.HasErrors() {
		validation.WriteResponse(w)
		return
	}

	//not an admin, exit
	if !validateAdmin(createImage.Organization, w, r) {
		return
	}

	revision, err := kiln.AutoRevision(createImage.Organization, createImage.Application, server.imageCreator)

	if err != nil {
		message := fmt.Sprintf("Unable to provide auto-revision to %s: %s", createImage.Application, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	baseImage, err := kiln.DetermineBaseImage(createImage.Runtime)

	if err != nil {
		message := fmt.Sprintf("Failed parsing runtime selection: %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	dockerInfo := &kiln.DockerInfo{
		RepoName:  createImage.Organization,
		ImageName: createImage.Application,
		EnvVars:   createImage.EnvVars,
		Revision:  revision,
		BaseImage: baseImage,
	}

	workspace, err := kiln.CreateNewWorkspace()

	if err != nil {
		message := fmt.Sprintf("Unable to create workspace, %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//point our source zip to the uploaded file

	//remove workspace after the request completes
	defer workspace.Clean()

	//copy the file data to a zip file

	//TODO REMOVE this copy
	//get the zip file and write bytes to it
	err = workspace.WriteZipFileData(file)

	if err != nil {
		message := fmt.Sprintf("Unable to write zip file %s", err)
		internalError(message, w)
		return
	}

	err = workspace.ExtractZipFile()

	if err != nil {
		message := fmt.Sprintf("Could not extract zip file %s ", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = workspace.CreateDockerFile(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Could not create docker file %s ", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = workspace.BuildTarFile()

	if err != nil {
		message := fmt.Sprintf("Could not create tar file %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	dockerBuild := &kiln.DockerBuild{
		DockerInfo: dockerInfo,
		TarFile:    workspace.TargetTarName,
	}

	outputChannel, err := server.imageCreator.BuildImage(dockerBuild)

	if err != nil {
		message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//write headers BEFORE setting status created
	w.Header().Set("Content-Type", "text/plain charset=utf-8")
	//turn this off so browsers render the response as it comes in
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Location", server.generateImageURL(dockerInfo, r.Host))
	//turn off proxy buffering in nginx (http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering)
	w.Header().Set("X-Accel-Buffering", "no")

	w.WriteHeader(http.StatusCreated)

	flusher, ok := w.(http.Flusher)

	//a serious problem, need to kill the server so we catch this during testing
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}

	//force chunked encoding
	flusher.Flush()

	//stream the log data
	err = chunkData(w, flusher, outputChannel)

	if err != nil {
		message := fmt.Sprintf("Could not flush data.  Error is %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//defer cleaning up the image
	// defer server.imageCreator.CleanImageRevision(dockerInfo)

	pushChannel, err := server.imageCreator.PushImage(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Could not push image from docker info %+v.  Error is %s", dockerInfo, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = chunkData(w, flusher, pushChannel)

	if err != nil {
		message := fmt.Sprintf("Could not flush data.  Error is %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	image, err := server.getImageInternal(createImage.Organization, createImage.Application, dockerInfo.Revision)

	if err != nil {
		message := fmt.Sprintf("Could not retrieve image for verification.  Error is %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	if image == nil {
		message := fmt.Sprintf("Unable to verify the image was pushed to the repository")
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	message := fmt.Sprintf("\nOrganization: %s | Application: %s | Revision: %s\n", dockerInfo.RepoName, dockerInfo.ImageName, dockerInfo.Revision)
	writeStringAndFlush(w, flusher, message)

}

//dup the data over to the response writer
func chunkData(w http.ResponseWriter, flusher http.Flusher, outputChannel chan (string)) error {

	kiln.LogInfo.Println("Beginning flushing of log data")

	for {

		data, ok := <-outputChannel

		kiln.LogInfo.Printf("Received data %s and ok %t", data, ok)

		if !ok {
			kiln.LogInfo.Printf("Received end of channel, breaking")
			break
		}

		err := writeStringAndFlush(w, flusher, data)

		if err != nil {
			return err
		}
	}

	kiln.LogInfo.Println("Completed flushing of log data")

	return nil

}

//write a string and flush it to the http Flushwriter
func writeStringAndFlush(w http.ResponseWriter, flusher http.Flusher, line string) error {

	_, err := w.Write([]byte(line))

	if err != nil {
		return err
	}

	flusher.Flush()

	return nil
}

//GetApplications returns an application
func (server *Server) getApplications(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	organization := vars["org"]

	//not an admin, exit
	if !validateAdmin(organization, w, r) {
		return
	}

	appNames, err := server.imageCreator.GetApplications(organization)

	if err != nil {
		message := fmt.Sprintf("Could not get images for organization %s.  Error is %s", organization, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	applications := []*Application{}

	for _, appName := range *appNames {
		application := &Application{
			Name: appName,
		}
		applications = append(applications, application)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

//getOrganizations get the organization
func (server *Server) getOrganizations(w http.ResponseWriter, r *http.Request) {

	//TODO, what's the security on this?  Open?  How can I validate they're an admin if I dont' see them, or do I filter?
	organizations := []*Organization{}

	organizationNames, err := server.imageCreator.GetOrganizations()

	if err != nil {
		message := fmt.Sprintf("Unable to retrieve organizations.  %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//get our token
	token, err := authsdk.NewJWTTokenFromRequest(r)

	if err != nil {
		message := fmt.Sprintf("Unable to find oAuth token %s", err)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return
	}

	// copy everything over
	for _, organization := range *organizationNames {

		kiln.LogInfo.Printf("Checking to see if user %s has admin authority for namepace %s", token.GetUsername(), organization)

		isAdmin, err := token.IsOrgAdmin(organization)

		if err != nil {
			message := fmt.Sprintf("Unable to get permission token %s", err)
			kiln.LogError.Printf(message)
			writeErrorResponse(http.StatusUnauthorized, message, w)
		}

		//if not an admin ignore this organization since theyr'e not allowed to see it
		if !isAdmin {
			continue
		}

		organizationObj := &Organization{
			Name: organization,
		}

		organizations = append(organizations, organizationObj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

//GetImages get the images for the given application
func (server *Server) getImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	organization := vars["org"]
	application := vars["name"]

	//not an admin, exit
	if !validateAdmin(organization, w, r) {
		return
	}

	dockerImages, err := server.imageCreator.GetImages(organization, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for organization %s and application %s.  Error is %s", organization, application, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	length := len(*dockerImages)

	//pre-allocate slice for efficiency
	images := make([]*Image, length)

	kiln.LogInfo.Printf("Processing %d images from the server", len(images))

	for i, image := range *dockerImages {

		resultImage := &Image{
			Revision: parseRevisionNumber(image.RepoTags[0]),
		}

		images[i] = resultImage
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

//GetImage get the image
func (server *Server) getImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	organization := vars["org"]
	application := vars["name"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(organization, w, r) {
		return
	}

	image, err := server.getImageInternal(organization, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for organization %s,  application %s, and revision %s.  Error is %s", organization, application, revision, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if image == nil {
		notFound(fmt.Sprintf("Could not get images for organization %s,  application %s, and revision %s.", organization, application, revision), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

//deleteApplication deletes all image revisions for an app
func (server *Server) deleteApplication(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	organization := vars["org"]
	application := vars["name"]

	//not an admin, exit
	if !validateAdmin(organization, w, r) {
		return
	}

	active, err := server.clusterConfig.CheckActiveDeployments(organization, application)
	if err != nil {
		message := fmt.Sprintf("Could not check deployments for organization %s and application %s.  Error is %s", organization, application, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	if active {
		message := fmt.Sprintf("Unable to delete application reivions for \"%s\" in \"%s\" because they are in use by an active deployment.\n", application, organization)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusConflict, message, w)
		return
	}

	dockerImages, err := server.imageCreator.GetImages(organization, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for organization %s and application %s.  Error is %s", organization, application, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if len(*dockerImages) == 0 {
		notFound(fmt.Sprintf("No images present for organization %s,  application %s.", organization, application), w)
		return
	}

	dockerInfo := &kiln.DockerInfo{
		RepoName:  organization,
		ImageName: application,
	}

	err = server.imageCreator.DeleteApplication(dockerInfo, dockerImages)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	//now write the response
	w.WriteHeader(http.StatusOK)
}

//getImageInternal get an image.  Image can be nil if not found, or an error will be returned if
func (server *Server) getImageInternal(organization string, application string, revision string) (*Image, error) {

	dockerInfo := &kiln.DockerInfo{
		RepoName:  organization,
		ImageName: application,
		Revision:  revision,
	}

	image, err := server.imageCreator.GetImageRevision(dockerInfo)

	if err != nil {
		return nil, err
	}

	if image == nil {
		return nil, nil
	}

	created := time.Unix(image.Created, 0)

	imageResponse := &Image{
		Revision: parseRevisionNumber(image.RepoTags[0]),
		ImageID:  image.ID,
		Created:  &created,
	}

	return imageResponse, nil
}

//generatePodSpec get the image
func (server *Server) generateImageURL(dockerInfo *kiln.DockerInfo, hostname string) string {

	endpoint := fmt.Sprintf("%s%s/%s/images/%s/version/%s", hostname, basePath, dockerInfo.RepoName, dockerInfo.ImageName, dockerInfo.Revision)

	return endpoint
}

//returns a 200 with "OK" body for health check
func (server *Server) status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func parseRevisionNumber(tag string) string {
	lastNdx := strings.LastIndex(tag, ":")
	return tag[lastNdx+1:]
}

//write a non 200 error response
func writeErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	w.WriteHeader(statusCode)

	errorObject := Error{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errorObject)
}

// internalError the error response when an internal error occurs
func internalError(message string, w http.ResponseWriter) {
	//log the error before we return it for debugging purposes
	writeErrorResponse(http.StatusInternalServerError, message, w)
}

// internalError the error response when an internal error occurs
func validationFailure(message string, w http.ResponseWriter) {
	//log the error before we return it for debugging purposes
	writeErrorResponse(http.StatusInternalServerError, message, w)
}

// notFound the error response when an internal error occurs
func notFound(message string, w http.ResponseWriter) {
	//log the error before we return it for debugging purposes
	writeErrorResponse(http.StatusNotFound, message, w)
}

//validateAdmin Validate the requestor is an admin in the namepace.  If returns false, the caller should halt and return.  True if the request should continue.  TODO make this cleaner
func validateAdmin(organization string, w http.ResponseWriter, r *http.Request) bool {

	//validate this user has a token and is org admin
	token, err := authsdk.NewJWTTokenFromRequest(r)

	if err != nil {
		message := fmt.Sprintf("Unable to find oAuth token %s", err)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	kiln.LogInfo.Printf("Checking to see if user %s has admin authority for namepace %s", token.GetUsername(), organization)

	isAdmin, err := token.IsOrgAdmin(organization)

	if err != nil {
		message := fmt.Sprintf("Unable to get permission token %s", err)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusForbidden, message, w)
		return false
	}

	//if not an admin, give access denied
	if !isAdmin {
		kiln.LogInfo.Printf("User %s is not an admin for organization %s", token.GetUsername(), organization)
		writeErrorResponse(http.StatusForbidden, fmt.Sprintf("You do not have admin permisison for organization %s", organization), w)
		return false
	}

	kiln.LogInfo.Printf("User %s is an admin for organization %s", token.GetUsername(), organization)

	return true
}
