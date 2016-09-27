package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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

const basePath = "/imagespaces"

const templateString = `BuildComplete
ID: %s
PodTemplateSpec: %s
`

//Server struct to create an instance of hte server
type Server struct {
	router       http.Handler
	decoder      *schema.Decoder
	imageCreator kiln.ImageCreator
	podSpecIo    kiln.PodspecIo
	template     *template.Template
}

//NewServer Create a new server using the provided podspecIo and Image creator.
func NewServer(imageCreator kiln.ImageCreator, podSpecIo kiln.PodspecIo) *Server {
	routes := mux.NewRouter()

	//allow the trailing slash
	//Note that when setting this to true, all URLS must end with a slash so we match paths both with and without the trailing slash
	// routes.StrictSlash(true)

	//now set up the decoder and return the server
	decoder := schema.NewDecoder()

	template := template.Must(template.New("outputTemplate").Parse(templateString))

	server := &Server{
		decoder:      decoder,
		imageCreator: imageCreator,
		podSpecIo:    podSpecIo,
		template:     template,
	}

	//a bit hacky, but need the pointer to the server
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/imagespaces/{org}/images/").HandlerFunc(server.postApplication)
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/imagespaces/{org}/images").HandlerFunc(server.postApplication)
	routes.Methods("GET").Path("/imagespaces/").HandlerFunc(server.getImagespaces) // getImagespaces
	routes.Methods("GET").Path("/imagespaces").HandlerFunc(server.getImagespaces)
	routes.Methods("GET").Path("/imagespaces/{org}/images/").HandlerFunc(server.getApplications) // get all images by name in imageSpace
	routes.Methods("GET").Path("/imagespaces/{org}/images").HandlerFunc(server.getApplications)
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}/").HandlerFunc(server.getImages) // get all revisions of an app in an imageSpace
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}").HandlerFunc(server.getImages)
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}/version/{revision}/").HandlerFunc(server.getImage) // get image by revision
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}/version/{revision}").HandlerFunc(server.getImage)
	routes.Methods("DELETE").Path("/imagespaces/{org}/images/{name}/version/{revision}/").HandlerFunc(server.deleteImage) // delete image by revision
	routes.Methods("DELETE").Path("/imagespaces/{org}/images/{name}/version/{revision}").HandlerFunc(server.deleteImage)

	//podtemplate generation
	routes.Methods("GET").Path("/imagespaces/generatepodspec/").Queries("imageURI", "", "publicPath", "").HandlerFunc(server.generatePodSpec)
	routes.Methods("GET").Path("/imagespaces/generatepodspec").Queries("imageURI", "", "publicPath", "").HandlerFunc(server.generatePodSpec)
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}/podspec/{revision}/").HandlerFunc(server.getPodSpec)
	routes.Methods("GET").Path("/imagespaces/{org}/images/{name}/podspec/{revision}").HandlerFunc(server.getPodSpec)
	//post a podspec for a specified revision
	routes.Methods("PUT").Headers("Content-Type", "application/json").Path("/imagespaces/{org}/images/{name}/podspec/{revision}/").HandlerFunc(server.postPodSpec)
	routes.Methods("PUT").Headers("Content-Type", "application/json").Path("/imagespaces/{org}/images/{name}/podspec/{revision}").HandlerFunc(server.postPodSpec)

	//health check
	routes.Methods("GET").Path("/imagespaces/status/").HandlerFunc(server.status)
	routes.Methods("GET").Path("/imagespaces/status").HandlerFunc(server.status)

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
	createImage.Imagespace = vars["org"]

	// r.PostForm is a map of our POST form values without the file
	err = server.decoder.Decode(createImage, r.Form)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	// Do something with person.Name or person.Phone

	validation := createImage.Validate()

	if validation.HasErrors() {
		validation.WriteResponse(w)
		return
	}

	//not an admin, exit
	if !validateAdmin(createImage.Imagespace, w, r) {
		return
	}

	dockerInfo := &kiln.DockerInfo{
		RepoName:  createImage.Imagespace,
		ImageName: createImage.Application,
		Revision:  createImage.Revision,
		EnvVars:   createImage.EnvVars,
		NodeVersion: createImage.NodeVersion,
	}

	//check if the image exists, if it does, return a 409

	existingImage, err := server.imageCreator.GetImageRevision(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Unable to check if image exists for %v", dockerInfo)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//image exists, don't allow the user to create it
	if existingImage != nil {
		writeErrorResponse(http.StatusConflict, fmt.Sprintf("An image in imageSpace %s with application %s and revision %s already exists", dockerInfo.RepoName, dockerInfo.ImageName, dockerInfo.Revision), w)
		return
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

	if os.Getenv("LOCAL_REGISTRY_ONLY") == "" {
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
	}

	image, err := server.getImageInternal(createImage.Imagespace, createImage.Application, createImage.Revision)

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

	//write the last portion
	outputString := fmt.Sprintf(templateString, image.ImageID, server.generatePodSpecURL(dockerInfo, r.Host, createImage.PublicPath))

	writeStringAndFlush(w, flusher, outputString)

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
	imageSpace := vars["org"]

	//not an admin, exit
	if !validateAdmin(imageSpace, w, r) {
		return
	}

	appNames, err := server.imageCreator.GetApplications(imageSpace)

	if err != nil {
		message := fmt.Sprintf("Could not get images for imageSpace %s.  Error is %s", imageSpace, err)
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

//getImagespaces get the imagespaces
func (server *Server) getImagespaces(w http.ResponseWriter, r *http.Request) {

	//TODO, what's the security on this?  Open?  How can I validate they're an admin if I dont' see them, or do I filter?
	imagespaces := []*Imagespace{}

	imagespaceNames, err := server.imageCreator.GetImagespaces()

	if err != nil {
		message := fmt.Sprintf("Unable to retrieve imagespaces.  %s", err)
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
	for _, imagespace := range *imagespaceNames {

		kiln.LogInfo.Printf("Checking to see if user %s has admin authority for namepace %s", token.GetUsername(), imagespace)

		isAdmin, err := token.IsOrgAdmin(imagespace)

		if err != nil {
			message := fmt.Sprintf("Unable to get permission token %s", err)
			kiln.LogError.Printf(message)
			writeErrorResponse(http.StatusUnauthorized, message, w)
		}

		//if not an admin ignore this imagespace since theyr'e not allowed to see it
		if !isAdmin {
			continue
		}

		imagespaceObj := &Imagespace{
			Name: imagespace,
		}

		imagespaces = append(imagespaces, imagespaceObj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imagespaces)
}

//GetImages get the images for the given application
func (server *Server) getImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageSpace := vars["org"]
	application := vars["name"]

	//not an admin, exit
	if !validateAdmin(imageSpace, w, r) {
		return
	}

	dockerImages, err := server.imageCreator.GetImages(imageSpace, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for imageSpace %s and application %s.  Error is %s", imageSpace, application, err)
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
			Created: time.Unix(image.Created, 0),
			ImageID: image.ID,
		}

		images[i] = resultImage
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

//GetImage get the image
func (server *Server) getImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageSpace := vars["org"]
	application := vars["name"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(imageSpace, w, r) {
		return
	}

	image, err := server.getImageInternal(imageSpace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for imageSpace %s,  application %s, and revision %s.  Error is %s", imageSpace, application, revision, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if image == nil {
		notFound(fmt.Sprintf("Could not get images for imageSpace %s,  application %s, and revision %s.", imageSpace, application, revision), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

//GetImage get the image
func (server *Server) deleteImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageSpace := vars["org"]
	application := vars["name"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(imageSpace, w, r) {
		return
	}

	image, err := server.getImageInternal(imageSpace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for imageSpace %s,  application %s, and revision %s.  Error is %s", imageSpace, application, revision, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if image == nil {
		notFound(fmt.Sprintf("Could not get images for imageSpace %s,  application %s, and revision %s.", imageSpace, application, revision), w)
		return
	}

	dockerInfo := &kiln.DockerInfo{
		RepoName:  imageSpace,
		ImageName: application,
		Revision:  revision,
	}

	err = server.imageCreator.DeleteImageRevision(dockerInfo)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	//now write the response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(image)
}

//getImageInternal get an image.  Image can be nil if not found, or an error will be returned if
func (server *Server) getImageInternal(imageSpace string, application string, revision string) (*Image, error) {

	dockerInfo := &kiln.DockerInfo{
		RepoName:  imageSpace,
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

	imageResponse := &Image{
		Created: time.Unix(image.Created, 0),
		ImageID: image.ID,
	}

	return imageResponse, nil
}

//postPodSpec get the image
func (server *Server) postPodSpec(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imagespace := vars["org"]
	application := vars["name"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(imagespace, w, r) {
		return
	}

	jsonBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		message := fmt.Sprintf("Could not read body. Error is %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	json := string(jsonBytes)

	//TODO validate

	//write an ok resposne
	err = server.podSpecIo.WritePodSpec(imagespace, application, revision, json)

	if err != nil {
		message := fmt.Sprintf("Could not write file. Error is %s", err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

//getPodSpec
func (server *Server) getPodSpec(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imagespace := vars["org"]
	name := vars["name"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(imagespace, w, r) {
		return
	}

	podSpec, err := server.podSpecIo.ReadPodSpec(imagespace, name, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get podspec for imagespace %s,  name %s, and revision %s.  Error is %s", imagespace, name, revision, err)
		kiln.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if podSpec == nil {
		notFound(fmt.Sprintf("Could not get podspec for imagespace %s,  name %s, and revision %s.", imagespace, name, revision), w)
		return
	}

	//write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(*podSpec))

}

//generatePodSpec get the image
func (server *Server) generatePodSpec(w http.ResponseWriter, r *http.Request) {

	//intentionally left open.
	queryParam := r.URL.Query()

	imageURI := queryParam.Get("imageURI")

	//validate the image uri is correct
	if imageURI == "" {
		internalError("You must specify a valid docker imageURI", w)
		return
	}

	//we purposefully don't validate these, since they're not required
	publicPath := queryParam.Get("publicPath")

	payload, err := kiln.GenerateKilnTemplateSpec(imageURI, publicPath)

	if err != nil {
		internalError(err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write([]byte(payload))

	if err != nil {
		internalError(err.Error(), w)
		return
	}
}

//generatePodSpec get the image
func (server *Server) generatePodSpecURL(dockerInfo *kiln.DockerInfo, hostname string, publicPath string) string {
	imageURI := server.imageCreator.GenerateRepoURI(dockerInfo)

	endpoint := fmt.Sprintf("https://%s%s/generatepodspec?imageURI=%s&publicPath=%s", hostname, basePath, imageURI, publicPath)

	return endpoint
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
func validateAdmin(imageSpace string, w http.ResponseWriter, r *http.Request) bool {

	//validate this user has a token and is org admin
	token, err := authsdk.NewJWTTokenFromRequest(r)

	if err != nil {
		message := fmt.Sprintf("Unable to find oAuth token %s", err)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	kiln.LogInfo.Printf("Checking to see if user %s has admin authority for namepace %s", token.GetUsername(), imageSpace)

	isAdmin, err := token.IsOrgAdmin(imageSpace)

	if err != nil {
		message := fmt.Sprintf("Unable to get permission token %s", err)
		kiln.LogError.Printf(message)
		writeErrorResponse(http.StatusForbidden, message, w)
		return false
	}

	//if not an admin, give access denied
	if !isAdmin {
		kiln.LogInfo.Printf("User %s is not an admin for imageSpace %s", token.GetUsername(), imageSpace)
		writeErrorResponse(http.StatusForbidden, fmt.Sprintf("You do not have admin permisison for imageSpace %s", imageSpace), w)
		return false
	}

	kiln.LogInfo.Printf("User %s is an admin for imageSpace %s", token.GetUsername(), imageSpace)

	return true
}
