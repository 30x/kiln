package kiln

//Interface for uploading

//PodspecIo the interface for reading and writing pod specs
type PodspecIo interface {
	//WritePodSpec write the pod spec
	WritePodSpec(namespace string, application string, revision string, podspec string) error

	//ReadPodSpec read the pod spec and return it as a string
	ReadPodSpec(namespace string, application string, revision string) (*string, error)
}
