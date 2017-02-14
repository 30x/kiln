package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

// GoogleContainerRegistry represents the Google Container Registry implementation of the remote Registry interface
type GoogleContainerRegistry struct {
	ProjectName string
	RegistryURL string
	Client      *http.Client
}

// NewGoogleContainerRegistryClient returns a newly configured Google Container Registry client
func NewGoogleContainerRegistryClient(registryURL string) (Registry, error) {
	ctx := context.Background()

	//Get our authn'd client
	client, err := google.DefaultClient(ctx)
	if err != nil {
		fmt.Printf("Error getting GCR client: %v\n", err)
		return nil, err
	}

	if registryURL == "" {
		registryURL = "https://gcr.io"
	}

	projectName, err := getGCPProjectName()
	if err != nil {
		return nil, err
	}

	return &GoogleContainerRegistry{
		ProjectName: projectName,
		RegistryURL: registryURL,
		Client:      client,
	}, nil
}

func getGCPProjectName() (string, error) {
	// curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id
	req, err := http.NewRequest(http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata-Flavor", "Google")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	data, _ := ioutil.ReadAll(res.Body)

	return string(data), nil
}

// GetProjectName returns the project name for this client
func (reg *GoogleContainerRegistry) GetProjectName() string {
	return reg.ProjectName
}

func (reg *GoogleContainerRegistry) getRegistryURLWithAPIVersionV2() string {
	return fmt.Sprintf("%s/v2", reg.RegistryURL)
}

// ListRepositories lists the image repos built under the given name
func (reg *GoogleContainerRegistry) ListRepositories(name string) ([]string, error) {
	// GCR /v2/_catalog support is weird, using their extended /tags/list functionality in the mean time
	target := fmt.Sprintf("%s/%s/%s/tags/list", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name)

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	repos := ListReposResponse{}
	err = decoder.Decode(&repos)
	if err != nil {
		return nil, err
	}

	return repos.Child, nil
}

// ListImageTags lists the available tags for an image
func (reg *GoogleContainerRegistry) ListImageTags(name string) ([]string, error) {
	target := fmt.Sprintf("%s/%s/%s/tags/list", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name)

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	tags := ListTagsResponse{}
	err = decoder.Decode(&tags)
	if err != nil {
		return nil, err
	}

	return tags.Tags, nil
}

// GetImageBlob retrieves an image blob specified by the name and reference
func (reg *GoogleContainerRegistry) GetImageBlob(name, reference string) (*ImageBlob, error) {
	target := fmt.Sprintf("%s/%s/%s/blobs/%s", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name, reference)

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	blob := ImageBlob{}
	err = decoder.Decode(&blob)
	if err != nil {
		return nil, err
	}

	return &blob, nil
}

// GetImageManifest retrieves a manifest for the image given by name and tag
func (reg *GoogleContainerRegistry) GetImageManifest(name, tag string) (*ManifestResponse, error) {
	target := fmt.Sprintf("%s/%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name, tag)

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	// need this header to get the correct digest
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := reg.Client.Do(req)
	if err != nil {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	manifest := ManifestResponse{}
	err = decoder.Decode(&manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

// DeleteImageManifest deletes the image manifest identified by the name and reference
// Note: the reference must be the manifest's digest, or it will only delete the tag
func (reg *GoogleContainerRegistry) DeleteImageManifest(name, reference string) error {
	target := fmt.Sprintf("%s/%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name, reference)

	req, err := http.NewRequest(http.MethodDelete, target, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := reg.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	return nil
}

// GetImageManifestDigest retrieves the image's manifest digest
func (reg *GoogleContainerRegistry) GetImageManifestDigest(name, tag string) (string, error) {
	target := fmt.Sprintf("%s/%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name, tag)

	req, err := http.NewRequest(http.MethodHead, target, nil)
	if err != nil {
		return "", err
	}

	// need this header to get the correct digest
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := reg.Client.Do(req)
	if err != nil {
		return "", nil
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	return res.Header.Get("Docker-Content-Digest"), nil
}

// GetImageBlobDigest retrieves the digest for the blob of the specified name and tag
func (reg *GoogleContainerRegistry) GetImageBlobDigest(name, tag string) (string, string, error) {
	target := fmt.Sprintf("%s/%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), reg.ProjectName, name, tag)

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "", "", err
	}

	// need this header to get the correct digest
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := reg.Client.Do(req)
	if err != nil {
		return "", "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Non-OK status code received: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	manifest := ManifestResponse{}
	err = decoder.Decode(&manifest)
	if err != nil {
		return "", "", err
	}

	return manifest.Config.Digest, res.Header.Get("Docker-Content-Digest"), nil
}
