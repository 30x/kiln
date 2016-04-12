package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/30x/shipyard/pkg/shipyard"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

//TODO make an env variable.  100 Meg max
const maxFileSize = 1024 * 1024 * 100

//Server struct to create an instance of hte server
type Server struct {
	router       *mux.Router
	decoder      *schema.Decoder
	imageCreator shipyard.ImageCreator
}

//NewServer Create a new server
func NewServer(imageCreator shipyard.ImageCreator) *Server {
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
	}

	//a bit hacky, but need the pointer to the server
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/images/").HandlerFunc(server.postApplication)
	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/images").HandlerFunc(server.postApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/").HandlerFunc(server.getNamespaces)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces").HandlerFunc(server.getNamespaces)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/").HandlerFunc(server.getApplications)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications").HandlerFunc(server.getApplications)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/").HandlerFunc(server.getApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}").HandlerFunc(server.getApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/images/").HandlerFunc(server.getImages)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/images").HandlerFunc(server.getImages)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/images/{revision}/").HandlerFunc(server.getImage)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namespace}/applications/{application}/images/{revision}").HandlerFunc(server.getImage)

	return server

}

//Start start the http server
func (server *Server) Start(port int) error {
	address := fmt.Sprintf(":%d", port)

	shipyard.LogInfo.Printf("Starting server at address %s", address)

	return http.ListenAndServe(address, server.router)
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

	dockerInfo := &shipyard.DockerInfo{
		RepoName:  createImage.Namespace,
		ImageName: createImage.Application,
		Revision:  createImage.Revision,
	}

	//check if the image exists, if it does, return a 409

	existingImage, err := server.imageCreator.GetImageRevision(dockerInfo.RepoName, dockerInfo.ImageName, dockerInfo.Revision)

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

	//

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

	//TODO make this a real writes
	// logWriter := &bytes.Buffer{}
	logWriter := os.Stdout

	err = server.imageCreator.BuildImage(dockerBuild, logWriter)

	if err != nil {
		message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = server.imageCreator.PushImage(dockerInfo, logWriter)

	if err != nil {
		message := fmt.Sprintf("Could not push image from docker info %+v.  Error is %s", dockerInfo, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	image := &Image{
		ImageID: dockerInfo.GetTagName(),
		Created: time.Now(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(image)

}

//GetApplications returns an application
func (server *Server) getApplications(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]

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

	json.NewEncoder(w).Encode(applications)
}

//getNamespaces get the namespaces
func (server *Server) getNamespaces(w http.ResponseWriter, r *http.Request) {

	namespaces := []*Namespace{}

	namespaceNames, err := server.imageCreator.GetRepositories()

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

	json.NewEncoder(w).Encode(namespaces)
}

//GetApplication return an application, if it exists
func (server *Server) getApplication(w http.ResponseWriter, r *http.Request) {

	//TODO finish this
	// vars := mux.Vars(r)
	// namespace := vars["namespace"]
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

	dockerImages, err := server.imageCreator.GetImages(namespace, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namespace %s and application %s.  Error is %s", namespace, application, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	images := []*Image{}

	for _, image := range *dockerImages {

		name := shipyard.GetImageNameFromTags(image.RepoTags)

		//can't parse it, skip it
		if name == nil {
			continue
		}

		image := &Image{
			Created: time.Unix(image.Created, 0),
			ImageID: *name,
			Size:    image.Size,
		}

		images = append(images, image)
	}

	json.NewEncoder(w).Encode(images)
}

//GetImage get the image
func (server *Server) getImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	application := vars["application"]
	revision := vars["revision"]

	image, err := server.imageCreator.GetImageRevision(namespace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namespace %s,  application %s, and revision %s.  Error is %s", namespace, application, revision, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	name := shipyard.GetImageNameFromTags(image.RepoTags)

	//not found, return a 404
	if name == nil {
		notFound(fmt.Sprintf("Could not get images for namespace %s,  application %s, and revision %s.", namespace, application, revision), w)
		return
	}

	imageResponse := &Image{
		Created: time.Unix(image.Created, 0),
		ImageID: *name,
		Size:    image.Size,
	}

	json.NewEncoder(w).Encode(imageResponse)
}

//write a non 200 error response
func writeErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	w.WriteHeader(statusCode)

	error := Error{
		Message: message,
	}

	json.NewEncoder(w).Encode(error)
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
