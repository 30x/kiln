package registry

import "time"

// ListTagsResponse represents the response from /tags/list
type ListTagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// CatalogResponse represents the reponse from a list of image repos
type CatalogResponse struct {
	Repositories []string `json:"repositories"`
}

// ListReposResponse here the child property represents the iamges built under the project
// this is based on GCR's extended /tags/list functionality
type ListReposResponse struct {
	Child    []string    `json:"child,omitempty"`
	Manifest interface{} `json:"manifest,omitempty"`
	Name     string      `json:"name"`
	Tags     []string    `json:"tags,omitempty"`
}

type ManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type ManifestLayer struct {
	MediaType string   `json:"mediaType"`
	Size      int      `json:"size"`
	Digest    string   `json:"digest"`
	URLs      []string `json:"urls,omitempty"`
}

type ManifestResponse struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        ManifestConfig  `json:"config"`
	Layers        []ManifestLayer `json:"layers"`
}

type ImageBlob struct {
	Architecture    string        `json:"architecture"`
	Config          BlobConfig    `json:"config"`
	Container       string        `json:"container"`
	ContainerConfig BlobConfig    `json:"container_config"`
	Created         time.Time     `json:"created"`
	DockerVersion   string        `json:"docker_version"`
	History         []BlobHistory `json:"history"`
	Os              string        `json:"os"`
	Rootfs          BlobRootFS    `json:"rootfs"`
}

type BlobConfig struct {
	ArgsEscaped  bool              `json:"ArgsEscaped"`
	AttachStderr bool              `json:"AttachStderr"`
	AttachStdin  bool              `json:"AttachStdin"`
	AttachStdout bool              `json:"AttachStdout"`
	Cmd          []string          `json:"Cmd"`
	Domainname   string            `json:"Domainname"`
	Entrypoint   interface{}       `json:"Entrypoint"`
	Env          []string          `json:"Env"`
	Hostname     string            `json:"Hostname"`
	Image        string            `json:"Image"`
	Labels       map[string]string `json:"Labels"`
	OnBuild      []interface{}     `json:"OnBuild"`
	OpenStdin    bool              `json:"OpenStdin"`
	StdinOnce    bool              `json:"StdinOnce"`
	Tty          bool              `json:"Tty"`
	User         string            `json:"User"`
	Volumes      interface{}       `json:"Volumes"`
	WorkingDir   string            `json:"WorkingDir"`
}

type BlobHistory struct {
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"created_by"`
	EmptyLayer bool      `json:"empty_layer,omitempty"`
}

type BlobRootFS struct {
	DiffIds []string `json:"diff_ids"`
	Type    string   `json:"type"`
}
