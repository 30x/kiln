package shipyard

import (
	"errors"
	"fmt"
	"strings"
)

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	RepoName  string
	ImageName string
	Revision  string
}

//GetImageName generate an image of the format {RepoName}/{ImageName}
func (dockerInfo *DockerInfo) GetImageName() string {
	return dockerInfo.RepoName + "/" + dockerInfo.ImageName
}

//GetTagName Get the anme of the tag of the format  {RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) GetTagName() string {
	return dockerInfo.GetImageName() + ":" + dockerInfo.Revision
}

//GetRemoteTagPath Get the remote tag name of the docker repo of the format {RemoteRepo}/{RepoName}/{ImageName}
func (dockerInfo *DockerInfo) GetRemoteTagPath(remoteRepo string) string {
	return remoteRepo + "/" + dockerInfo.GetImageName()
}

//GetRemoteTagName Get the remote tag name of the docker repo of the format {RemoteRepo}/{RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) GetRemoteTagName(remoteRepo string) string {
	return remoteRepo + "/" + dockerInfo.GetTagName()
}

//DockerBuild a type for building a docker image docker
type DockerBuild struct {
	TarFile string
	*DockerInfo
}

//RepoImage an extension of a local image with a remote repository
type RepoImage struct {
	//registry URI
	Registry string
	//OriginalURI the original URI passed to the user
	OriginalURI string
	DockerInfo
}

//NewRepoImage Parse the repo URI
func NewRepoImage(URI string) (*RepoImage, error) {
	parts := strings.Split(URI, "/")

	if len(parts) != 3 {
		return nil, errors.New("You must supply an absolute URI.  The format should be {repo name}/{registry}/{image}:{tag}")
	}

	tags := strings.Split(parts[2], ":")

	if len(tags) != 2 {
		return nil, errors.New("You must supply an absolute URI.  The format should be {repo name}/{registry}/{image}:{tag}")
	}

	return &RepoImage{
		OriginalURI: URI,
		Registry:    parts[0],
		DockerInfo: DockerInfo{
			RepoName:  parts[1],
			ImageName: tags[0],
			Revision:  tags[1]},
	}, nil
}

//GenerateName Generate a unqiue app name from this repo image
func (repoImage *RepoImage) GenerateName() string {
	return repoImage.OriginalURI
}

//GeneratePodName Get the pod name of the format {RepoName}_{ImageName}_{Revision}
func (repoImage *RepoImage) GeneratePodName() string {
	return fmt.Sprintf("%s-%s-%s", repoImage.RepoName, repoImage.ImageName, repoImage.Revision)
}
