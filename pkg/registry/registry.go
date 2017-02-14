package registry

// Registry generic interface for interacting with a Docker Registry
type Registry interface {

	// ListRepositories lists the image repositories under the given name (usually a registry/project name)
	ListRepositories(name string) ([]string, error)

	// ListImageTags lists all available image tags for the given name
	ListImageTags(name string) ([]string, error)

	// GetImageBlobDigest retrieves an image's unique blob digest, as well as the manifest digest b.c we get that for free, based on the name and tag
	GetImageBlobDigest(name, tag string) (string, string, error)

	// GetImageManifestDigest retrieves an image's unique manifest digest based on the name and tag
	GetImageManifestDigest(name, tag string) (string, error)

	// GetImageBlob retrieves the blob (details) for the image specified by name and reference
	GetImageBlob(name, reference string) (*ImageBlob, error)

	// GetImageManifest retrieves the manifest for the given image name & tag
	GetImageManifest(name, tag string) (*ManifestResponse, error)

	// DeleteImageManifest deletes the image identified by the name and tag
	DeleteImageManifest(name, reference string) error

	// GetProjectName returns the GCR Project Name, private registry doesn't have this property
	GetProjectName() string
}
