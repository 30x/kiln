package shipyard

//DockerInfo is a struct that holds information for creating a docker container
type DockerInfo struct {
	RepoName  string
	ImageName string
	Revision  string
}

//getImageName generate an image of the format {RepoName}/{ImageName}
func (dockerInfo *DockerInfo) GetImageName() string {
	return dockerInfo.RepoName + "/" + dockerInfo.ImageName
}

//getTagName Get the anme of the tag of the format  {RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) GetTagName() string {
	return dockerInfo.GetImageName() + ":" + dockerInfo.Revision
}

//getRemoteTagName Get the remote tag name of the docker repo of the format {RemoteRepo}/{RepoName}/{ImageName}:{Revision}
func (dockerInfo *DockerInfo) GetRemoteTagName(remoteRepo string) string {
	return remoteRepo + "/" + dockerInfo.GetImageName()
}

//DockerBuild a type for building a docker image docker
type DockerBuild struct {
	TarFile string
	*DockerInfo
}

//ImageSearch A type for performing searches
type ImageSearch struct {
	Repository  string
	Application string
	Revision    string
}
