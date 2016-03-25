package server

// workspace, err := shipyard.CreateNewWorkspace()

// 		if err != nil {
// 			message := fmt.Sprintf("Unable to create workspace, %s", err)
// 			return InternalError(message)
// 		}

// 		//remove workspace after the request completes
// 		defer workspace.Clean()

// 		//copy the file data to a zip file
// 		base64FileData := params.File.Data

// 		// byteData := []byte{}

// 		// err = base64FileData.UnmarshalText(byteData)

// 		// _, err = base64.URLEncoding.Decode(byteData, []byte(base64FileData))

// 		// if err != nil {
// 		// 	message := fmt.Sprintf("Unable to unmarshall base64 into bytes %s", err)
// 		// 	return InternalError(message)
// 		// }

// 		//get the zip file and write bytes to it
// 		err = workspace.WriteZipeFileData(base64FileData)

// 		if err != nil {
// 			message := fmt.Sprintf("Unable to write zip file %s", err)
// 			return InternalError(message)
// 		}

// 		dockerInfo := &shipyard.DockerInfo{
// 			RepoName:  params.Repository,
// 			ImageName: params.Application,
// 			Revision:  params.Revision,
// 		}

// 		dockerFile := &shipyard.DockerFile{
// 			ParentImage: "node:4.3.0-onbuild",
// 			DockerInfo:  dockerInfo,
// 		}

// 		err = workspace.CreateDockerFile(dockerFile)

// 		if err != nil {
// 			message := fmt.Sprintf("Could not create docker file %s ", err)
// 			return InternalError(message)
// 		}

// 		err = workspace.BuildTarFile()

// 		if err != nil {
// 			message := fmt.Sprintf("Could not create tar file %s", err)
// 			return InternalError(message)
// 		}

// 		dockerBuild := &shipyard.DockerBuild{
// 			DockerInfo: dockerInfo,
// 			TarFile:    workspace.TargetTarName,
// 		}

// 		//TODO make this a real writes
// 		logWriter := os.Stdout

// 		err = imageCreator.BuildImage(dockerBuild, logWriter)

// 		if err != nil {
// 			message := fmt.Sprintf("Could not build image from docker info %+v.  Error is %s", dockerInfo, err)
// 			return InternalError(message)
// 		}

// 		response := operations.NewCreateApplicationCreated()

// 		response.Payload = &models.Image{
// 			Created: strfmt.NewDateTime(),
// 			ImageID: dockerInfo.GetImageName(),
// 		}

// 		return response
// 	})

// 	api.GetAllApplicationsHandler = operations.GetAllApplicationsHandlerFunc(func(params operations.GetAllApplicationsParams) middleware.Responder {

// 		dockerInfo := &shipyard.DockerInfo{
// 			RepoName: params.Repository,
// 		}

// 		images, err := imageCreator.SearchRemoteImages(dockerInfo)

// 		if err != nil {
// 			message := fmt.Sprintf("Could not search docker images %+v.  Error is %s", dockerInfo, err)
// 			return InternalError(message)
// 		}

// 		response := operations.NewGetAllApplicationsOK()

// 		for _, image := range images {

// 			if image.RepoTags == nil || len(image.RepoTags) == 0 {
// 				continue
// 			}

// 			application := &models.Application{
// 				Name: image.RepoTags[0],
// 			}

// 			response.Payload = append(response.Payload, application)
// 		}

// 		return response
