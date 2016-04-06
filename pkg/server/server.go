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

const dockerRegistryVar = "DOCKER_REGISTRY_URL"

//Server struct to create an instance of hte server
type Server struct {
	router  *mux.Router
	decoder *schema.Decoder
}

//NewServer Create a new server
func NewServer() (server *Server) {
	r := mux.NewRouter()
	routes := r.PathPrefix("/beeswax/images/api/v1").Subrouter()

	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/namespaces/{namspace}/applications").HandlerFunc(postApplication)
	// routes.Methods("POST").Path("/{namspace}/applications").HandlerFunc(postApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces").HandlerFunc(getNamespaces)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namspace}/applications").HandlerFunc(getApplications)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namspace}/applications/{application}").HandlerFunc(getApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namspace}/applications/{application}/images").HandlerFunc(getImages)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/namespaces/{namspace}/applications/{application}/images/{revision}").HandlerFunc(getImage)

	server = &Server{
		router: r,
	}

	return server

}

var decoder *schema.Decoder
var imageCreator shipyard.ImageCreator

func init() {
	decoder = schema.NewDecoder()

	var error error

	repoURL := os.Getenv(dockerRegistryVar)

	if repoURL == "" {
		shipyard.LogError.Fatalf("You must set the %s environment variable.  An example would be localhost:5000", dockerRegistryVar)
	}

	// imageCreator, error = shipyard.NewEcsImageCreator("977777657611.dkr.ecr.us-east-1.amazonaws.com", "us-east-1")
	imageCreator, error = shipyard.NewLocalImageCreator(repoURL)

	//we should die here if we're unable to start
	if error != nil {
		shipyard.LogError.Fatalf("Unable to create image creator %s", error)
	}

}

//Start start the http server
func (server *Server) Start(port int) error {
	address := fmt.Sprintf(":%d", port)

	shipyard.LogInfo.Printf("Starting server at address %s", address)

	return http.ListenAndServe(address, server.router)
}

//postApplication and render a response
func postApplication(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namspace := vars["namspace"]

	err := r.ParseMultipartForm(maxFileSize)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	createApplication := new(CreateApplication)

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

	// r.PostForm is a map of our POST form values without the file
	err = decoder.Decode(createApplication, r.Form)

	if err != nil {
		message := fmt.Sprintf("Unable parse form %s", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	// Do something with person.Name or person.Phone

	validation := createApplication.Validate()

	if validation.HasErrors() {
		validation.WriteResponse(w)
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

	dockerInfo := &shipyard.DockerInfo{
		RepoName:  namspace,
		ImageName: createApplication.Application,
		Revision:  createApplication.Revision,
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

	err = imageCreator.BuildImage(dockerBuild, logWriter)

	if err != nil {
		message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = imageCreator.PushImage(dockerInfo, logWriter)

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

	json.NewEncoder(w).Encode(image)

}

//GetApplications returns an application
func getApplications(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namspace := vars["namspace"]

	appNames, err := imageCreator.GetApplications(namspace)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namspace %s.  Error is %s", namspace, err)
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
func getNamespaces(w http.ResponseWriter, r *http.Request) {

	namespaces := []*Namespace{}

	json.NewEncoder(w).Encode(namespaces)
}

//GetApplication return an application, if it exists
func getApplication(w http.ResponseWriter, r *http.Request) {

	//TODO finish this
	// vars := mux.Vars(r)
	// namspace := vars["namspace"]
	// application := vars["application"]

	// appNames, err := imageCreator.GetApplications(namspace)

	// if err != nil {
	// 	message := fmt.Sprintf("Could not get images for namspace %s.  Error is %s", namspace, err)
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
func getImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namspace := vars["namspace"]
	application := vars["application"]

	dockerImages, err := imageCreator.GetImages(namspace, application)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namspace %s and application %s.  Error is %s", namspace, application, err)
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
func getImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	namspace := vars["namspace"]
	application := vars["application"]
	revision := vars["revision"]

	image, err := imageCreator.GetImageRevision(namspace, application, revision)

	if err != nil {
		message := fmt.Sprintf("Could not get images for namspace %s,  application %s, and revision %s.  Error is %s", namspace, application, revision, err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	name := shipyard.GetImageNameFromTags(image.RepoTags)

	//not found, return a 404
	if name == nil {
		notFound(fmt.Sprintf("Could not get images for namspace %s,  application %s, and revision %s.", namspace, application, revision), w)
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
