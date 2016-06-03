package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/30x/authsdk"
	"github.com/30x/shipyard/pkg/shipyard"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/tylerb/graceful"
)

//TODO make an env variable.  100 Meg max
const maxFileSize = 1024 * 1024 * 100

//Server struct to create an instance of hte server
type Server struct {
	router       *mux.Router
	decoder      *schema.Decoder
	imageCreator shipyard.ImageCreator
	podSpecIo    shipyard.PodspecIo
}

//NewServer Create a new server
func NewServer(imageCreator shipyard.ImageCreator, podSpecIo shipyard.PodspecIo) *Server {
	r := mux.NewRouter()

	routes := r.PathPrefix("/beeswax/images/api/v1").Subrouter()

	//allow the trailing slash
	//Note that when setting this to true, all URLS must end with a slash so we match paths both with and without the trailing slash
	// routes.StrictSlash(true)

	//now set up the decoder and return the server
	decoder := schema.NewDecoder()

	server := &Server{
		router:       r,
		decoder:      decoder,
		imageCreator: imageCreator,
		podSpecIo:    podSpecIo,
	}

	//a bit hacky, but need the pointer to the server
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/builds/").HandlerFunc(server.postApplication)
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/builds").HandlerFunc(server.postApplication)
	routes.Methods("GET").Path("/namespaces/").HandlerFunc(server.getNamespaces)
	routes.Methods("GET").Path("/namespaces").HandlerFunc(server.getNamespaces)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/").HandlerFunc(server.getApplications)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications").HandlerFunc(server.getApplications)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/").HandlerFunc(server.getApplication)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}").HandlerFunc(server.getApplication)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/images/").HandlerFunc(server.getImages)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/images").HandlerFunc(server.getImages)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/images/{revision}/").HandlerFunc(server.getImage)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/images/{revision}").HandlerFunc(server.getImage)

	//post a podspec for a specified revision
	routes.Methods("PUT").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/podspec/{revision}/").HandlerFunc(server.postPodSpec)
	routes.Methods("PUT").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/podspec/{revision}").HandlerFunc(server.postPodSpec)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/podspec/{revision}/").HandlerFunc(server.getPodSpec)
	routes.Methods("GET").Path("/namespaces/{namespace}/applications/{application}/podspec/{revision}").HandlerFunc(server.getPodSpec)

	//podtemplate generation
	routes.Methods("GET").Path("/generatepodspec/").Queries("imageURI", "", "publicPath", "").HandlerFunc(server.generatePodSpec)
	routes.Methods("GET").Path("/generatepodspec").Queries("imageURI", "", "publicPath", "").HandlerFunc(server.generatePodSpec)

	return server

}

//Start start the http server with the port and the specified timeout, as well as the shutdown timer
func (server *Server) Start(port int, timeout time.Duration, shutdownTimer time.Duration) {
	address := fmt.Sprintf(":%d", port)

	shipyard.LogInfo.Printf("Starting server at address %s", address)

	srv := &graceful.Server{
		Timeout: timeout,
		Server:  &http.Server{Addr: address, Handler: server.router},
		Logger:  graceful.DefaultLogger(),
	}

	//set up our timer in a gofunc in order to shut down after a duration

	go func() {

		//do nothing, we dont' want to shut down
		if time.Duration(0) == shutdownTimer {
			return
		}

		seed := time.Now().Unix()
		rand.Seed(seed)

		//add a 10 minute variance to our shutdown timer
		additionalSeconds := rand.Intn(60)

		shipyard.LogInfo.Printf("additionalSeconds %d", additionalSeconds)

		additionalSecondsDuration := time.Duration(additionalSeconds) * time.Second

		shipyard.LogInfo.Printf("additionalSecondsDuration %f", additionalSecondsDuration.Seconds())

		total := shutdownTimer + additionalSecondsDuration

		shipyard.LogInfo.Printf("The system will shut down in %f seconds", total.Seconds())

		timer := time.NewTimer(total)
		<-timer.C

		//shut down, we've hit our timer
		os.Exit(0)
	}()

	//start listening
	if err := srv.ListenAndServe(); err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			srv.Logger.Fatal(err)
		}
	}

}

//postApplication and render a response
func (server *Server) postApplication(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	// file, handler, err := r.FormFile("file")
	file, _, err := r.FormFile("file")

	if err != nil {
		message := fmt.Sprintf("Unable to upload file %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//defer closing after request completes
	defer file.Close()

	createImage := new(CreateImage)

	// r.PostForm is a map of our POST form values without the file
	err = server.decoder.Decode(createImage, r.Form)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		shipyard.LogError.Printf(message)
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
	if !validateAdmin(createImage.Namespace, w, r) {
		return
	}

	dockerInfo := &shipyard.DockerInfo{
		RepoName:  createImage.Namespace,
		ImageName: createImage.Application,
		Revision:  createImage.Revision,
	}

	//check if the image exists, if it does, return a 409

	existingImage, err := server.imageCreator.GetImageRevision(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Unable to check if image exists for %v", dockerInfo)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//image exists, don't allow the user to create it
	if existingImage != nil {
		writeErrorResponse(http.StatusConflict, fmt.Sprintf("An image in namespace %s with application %s and revision %s already exists", dockerInfo.RepoName, dockerInfo.ImageName, dockerInfo.Revision), w)
		return
	}

	workspace, err := shipyard.CreateNewWorkspace()

	if err != nil {
		message := fmt.Sprintf("Unable to create workspace, %s", err)
		shipyard.LogError.Printf(message)
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
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = workspace.CreateDockerFile(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Could not create docker file %s ", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = workspace.BuildTarFile()

	if err != nil {
		message := fmt.Sprintf("Could not create tar file %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	dockerBuild := &shipyard.DockerBuild{
		DockerInfo: dockerInfo,
		TarFile:    workspace.TargetTarName,
	}

	outputChannel, err := server.imageCreator.BuildImage(dockerBuild)

	if err != nil {
		message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//write headers BEFORE setting status created
	w.Header().Set("Content-Type", "text/plain charset=utf-8")
	//turn this off so browsers render the response as it comes in
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(http.StatusCreated)

	flusher, ok := w.(http.Flusher)

	//force chunked encoding
	flusher.Flush()

	//a serious problem, need to kill the server so we catch this during testing
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}

	//stream the log data
	err = chunkData(w, flusher, outputChannel)

	if err != nil {
		message := fmt.Sprintf("Could not flush data.  Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//defer cleaning up the image
	defer server.imageCreator.CleanImageRevision(dockerInfo)

	pushChannel, err := server.imageCreator.PushImage(dockerInfo)

	if err != nil {
		message := fmt.Sprintf("Could not push image from docker info %+v.  Error is %s", dockerInfo, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = chunkData(w, flusher, pushChannel)

	if err != nil {
		message := fmt.Sprintf("Could not flush data.  Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	image, err := server.getImageInternal(createImage.Namespace, createImage.Application, createImage.Revision)

	if err != nil {
		message := fmt.Sprintf("Could not retrieve image for verification.  Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	if image == nil {
		message := fmt.Sprintf("Unable to verify the image was pushed to the repository")
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}
	//write the last portion

	finalOutput := fmt.Sprintf("\nBuild Complete \n%s", image.ImageID)

	err = writeStringAndFlush(w, flusher, finalOutput)

	if err != nil {
		message := fmt.Sprintf("Could not flush data.  Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

}

//dup the data over to the response writer
func chunkData(w http.ResponseWriter, flusher http.Flusher, outputChannel chan (string)) error {

	shipyard.LogInfo.Println("Beginning flushing of log data")

	for {

		data, ok := <-outputChannel

		shipyard.LogInfo.Printf("Received data %s and ok %t", data, ok)

		if !ok {
			shipyard.LogInfo.Printf("Received end of channel, breaking")
			break
		}

		err := writeStringAndFlush(w, flusher, data)

		if err != nil {
			return err
		}
	}

	shipyard.LogInfo.Println("Completed flushing of log data")

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
	namespace := vars["namespace"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	appNames, err := server.imageCreator.GetApplications(namespace)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namespace %s.  Error is %s", namespace, err)
		shipyard.LogError.Printf(message)
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

//getNamespaces get the namespaces
func (server *Server) getNamespaces(w http.ResponseWriter, r *http.Request) {

	//TODO, what's the security on this?  Open?  How can I validate they're an admin if I dont' see them, or do I filter?
	namespaces := []*Namespace{}

	namespaceNames, err := server.imageCreator.GetNamespaces()

	if err != nil {
		message := fmt.Sprintf("Unable to retrieve namespaces.  %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	// copy everything over
	for _, namespace := range *namespaceNames {
		namespaceObj := &Namespace{
			Name: namespace,
		}

		namespaces = append(namespaces, namespaceObj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(namespaces)
}

//GetApplication return an application, if it exists
func (server *Server) getApplication(w http.ResponseWriter, r *http.Request) {

	//TODO finish this
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	// application := vars["application"]

	// appNames, err := imageCreator.GetApplications(namespace)

	// if err != nil {
	// 	message := fmt.Sprintf("Could not get images for namespace %s.  Error is %s", namespace, err)
	// 	shipyard.LogError.Printf(message)
	// 	internalError(message, w)
	// 	return
	// }

	// applications := []*Application{}

	// for _, appName := range *appNames {
	// 	application := &Application{
	// 		Name: appName,
	// 	}
	// 	applications = append(applications, application)
	// }

	// json.NewEncoder(w).Encode(applications)
}

//GetImages get the images for the given application
func (server *Server) getImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	application := vars["application"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	dockerImages, err := server.imageCreator.GetImages(namespace, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namespace %s and application %s.  Error is %s", namespace, application, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	length := len(*dockerImages)

	//pre-allocate slice for efficiency
	images := make([]*Image, length)

	shipyard.LogInfo.Printf("Processing %d images from the server", len(images))

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
	namespace := vars["namespace"]
	application := vars["application"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	image, err := server.getImageInternal(namespace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namespace %s,  application %s, and revision %s.  Error is %s", namespace, application, revision, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if image == nil {
		notFound(fmt.Sprintf("Could not get images for namespace %s,  application %s, and revision %s.", namespace, application, revision), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

//getImageInternal get an image.  Image can be nil if not found, or an error will be returned if
func (server *Server) getImageInternal(namespace string, application string, revision string) (*Image, error) {

	dockerInfo := &shipyard.DockerInfo{
		RepoName:  namespace,
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
	namespace := vars["namespace"]
	application := vars["application"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	jsonBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		message := fmt.Sprintf("Could not read body. Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	json := string(jsonBytes)

	//TODO validate

	//write an ok resposne
	err = server.podSpecIo.WritePodSpec(namespace, application, revision, json)

	if err != nil {
		message := fmt.Sprintf("Could not write file. Error is %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

//getPodSpec
func (server *Server) getPodSpec(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	application := vars["application"]
	revision := vars["revision"]

	//not an admin, exit
	if !validateAdmin(namespace, w, r) {
		return
	}

	podSpec, err := server.podSpecIo.ReadPodSpec(namespace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get podspec for namespace %s,  application %s, and revision %s.  Error is %s", namespace, application, revision, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	//not found, return a 404
	if podSpec == nil {
		notFound(fmt.Sprintf("Could not get podspec for namespace %s,  application %s, and revision %s.", namespace, application, revision), w)
		return
	}

	//write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(*podSpec))

}

//GetImage get the image
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

	payload, err := shipyard.GenerateShipyardTemplateSpec(imageURI, publicPath)

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
func validateAdmin(namespace string, w http.ResponseWriter, r *http.Request) bool {

	//validate this user has a token and is org admin
	token, err := authsdk.NewJWTTokenFromRequest(r)

	if err != nil {
		message := fmt.Sprintf("Unable to find oAuth token %s", err)
		shipyard.LogError.Printf(message)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	shipyard.LogInfo.Printf("Checking to see if user %s has admin authority for namepace %s", token.GetUsername(), namespace)

	isAdmin, err := token.IsOrgAdmin(namespace)

	if err != nil {
		message := fmt.Sprintf("Unable to get permission token %s", err)
		shipyard.LogError.Printf(message)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	//if not an admin, give access denied
	if !isAdmin {
		shipyard.LogInfo.Printf("User %s is not an admin for namespace %s", token.GetUsername(), namespace)
		writeErrorResponse(http.StatusForbidden, fmt.Sprintf("You do not have admin permisison for namespace %s", namespace), w)
		return false
	}

	shipyard.LogInfo.Printf("User %s is an admin for namespace %s", token.GetUsername(), namespace)

	return true
}
