package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// PrivateRegistry represents a private Docker Registry client
type PrivateRegistry struct {
	RegistryURL string
	Client      *http.Client
	AuthScheme  string
	AuthToken   string
}

// NewPrivateRegistryClient creates a client for use with a private Docker Registry
func NewPrivateRegistryClient(registryURL string) Registry {
	reg := PrivateRegistry{}

	// if these are empty, they won't be used later on
	// TODO: figure out how to use golang.org/x/oauth2 to build client
	reg.AuthScheme = os.Getenv("REGISTRY_AUTH_SCHEME")
	reg.AuthToken = os.Getenv("REGISTRY_AUTH_TOKEN")

	reg.Client = http.DefaultClient
	reg.RegistryURL = registryURL

	return &reg
}

func (reg *PrivateRegistry) getRegistryURLWithAPIVersionV2() string {
	return fmt.Sprintf("%s/v2", reg.RegistryURL)
}

func (reg *PrivateRegistry) checkAndAddAuth(req *http.Request) {
	if reg.AuthScheme != "" && reg.AuthToken != "" {
		req.Header.Add("Authorization", reg.AuthScheme+" "+reg.AuthToken)
	}
}

// ListRepositories lists the image repos built under the given name
func (reg *PrivateRegistry) ListRepositories(name string) ([]string, error) {
	target := fmt.Sprintf("%s/_catalog", reg.getRegistryURLWithAPIVersionV2())

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received in ListRepositories: %d", res.StatusCode)
	}

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	repos := CatalogResponse{}
	err = decoder.Decode(&repos)
	if err != nil {
		return nil, err
	}

	if name != "" {
		// need to filter results by name
		return filterReposByRootName(repos.Repositories, name), nil
	}

	return repos.Repositories, nil
}

// ListImageTags lists the available tags for an image
func (reg *PrivateRegistry) ListImageTags(name string) ([]string, error) {
	target := fmt.Sprintf("%s/%s/tags/list", reg.getRegistryURLWithAPIVersionV2(), name)

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received in ListImageTags: %d", res.StatusCode)
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
func (reg *PrivateRegistry) GetImageBlob(name, reference string) (*ImageBlob, error) {
	target := fmt.Sprintf("%s/%s/blobs/%s", reg.getRegistryURLWithAPIVersionV2(), name, reference)

	res, err := reg.Client.Get(target)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code received in GetImageBlob: %d", res.StatusCode)
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
func (reg *PrivateRegistry) GetImageManifest(name, tag string) (*ManifestResponse, error) {
	target := fmt.Sprintf("%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), name, tag)

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
		return nil, fmt.Errorf("Non-OK status code received in GetImageManifest: %d", res.StatusCode)
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
func (reg *PrivateRegistry) DeleteImageManifest(name, reference string) error {
	target := fmt.Sprintf("%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), name, reference)

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
		return fmt.Errorf("Non-OK status code received in DeleteImageManifest: %d", res.StatusCode)
	}

	return nil
}

// GetImageManifestDigest retrieves the image's manifest digest
func (reg *PrivateRegistry) GetImageManifestDigest(name, tag string) (string, error) {
	target := fmt.Sprintf("%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), name, tag)

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
		return "", fmt.Errorf("Non-OK status code received in GetImageManifestDigest: %d", res.StatusCode)
	}

	return res.Header.Get("Docker-Content-Digest"), nil
}

// GetImageBlobDigest retrieves the digest for the blob of the specified name and tag
func (reg *PrivateRegistry) GetImageBlobDigest(name, tag string) (string, string, error) {
	target := fmt.Sprintf("%s/%s/manifests/%s", reg.getRegistryURLWithAPIVersionV2(), name, tag)

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
		return "", "", fmt.Errorf("Non-OK status code received in GetImageBlobDigest: %d", res.StatusCode)
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

// GetProjectName returns an empty string b.c this is only needed for GCR
func (reg *PrivateRegistry) GetProjectName() string {
	return ""
}

// DeleteImageTag deletes the image's specified tag
func (reg *PrivateRegistry) DeleteImageTag(name, tag string) error {
	return fmt.Errorf("Deleting an image manfiest by tag is not supported in a private registry")
}
