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

	routes.Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Path("/{repository}/applications").HandlerFunc(postApplication)
	// routes.Methods("POST").Path("/{repository}/applications").HandlerFunc(postApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/{repository}/applications").HandlerFunc(GetApplications)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/{repository}/applications/{application}").HandlerFunc(GetApplication)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/{repository}/applications/{application}/images").HandlerFunc(GetImages)
	routes.Methods("GET").Headers("Content-Type", "application/json").Path("/{repository}/applications/{application}/images/{revision}").HandlerFunc(GetImage)

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
	repository := vars["repository"]

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

	// byteData := []byte{}

	// size, err := base64.URLEncoding.Decode(byteData, []byte(createApplication.File))

	// if err != nil {
	// 	message := fmt.Sprintf("Unable to unmarshall base64 into %d bytes %s", size, err)
	// 	internalError(message, w)
	// 	return
	// }

	//TODO REMOVE this copy
	//get the zip file and write bytes to it
	err = workspace.WriteZipFileData(file)

	if err != nil {
		message := fmt.Sprintf("Unable to write zip file %s", err)
		internalError(message, w)
		return
	}

	dockerInfo := &shipyard.DockerInfo{
		RepoName:  repository,
		ImageName: createApplication.Application,
		Revision:  createApplication.Revision,
	}

	dockerFile := &shipyard.DockerFile{
		ParentImage: "node:4.3.0-onbuild",
		DockerInfo:  dockerInfo,
	}

	err = workspace.ExtractZipFile()

	if err != nil {
		message := fmt.Sprintf("Could not extract zip file %s ", err)
		shipyard.LogError.Printf(message)
		internalError(message, w)
		return
	}

	err = workspace.CreateDockerFile(dockerFile)

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

func GetApplications(http.ResponseWriter, *http.Request) {

}

func GetApplication(http.ResponseWriter, *http.Request) {

}

func GetImages(http.ResponseWriter, *http.Request) {

}

func GetImage(http.ResponseWriter, *http.Request) {

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
