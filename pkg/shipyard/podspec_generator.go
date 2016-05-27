package shipyard

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"
)

//GenerateShipyardTemplateSpec generate a valid pod template spec given the docker image uri, and return it as a json encoded string
func GenerateShipyardTemplateSpec(dockerURI string, publicPath string) (string, error) {

	//TODO validate we only ever have 1 port in the public paths.  Parse out the port and then set it below.

	repoImage, err := NewRepoImage(dockerURI)

	if err != nil {
		return "", err
	}

	//validate the public path

	parts := strings.Split(publicPath, ":")

	port := ""

	//for shipyard, we should only have 1 port and path
	if len(parts) != 2 {
		return "", errors.New("Only 1 public path is supported. It must be of the format {PORT}:/{PATH?}")
	}

	port = parts[0]

	intPort, err := strconv.Atoi(port)

	if err != nil {
		return "", errors.New("Port must be parsable to an int")
	}

	podTemplate := api.PodTemplateSpec{
		ObjectMeta: api.ObjectMeta{
			Labels: map[string]string{
				"runtime":  "shipyard",
				"app":      repoImage.GeneratePodName(),
				"routable": "true",
			},
			Annotations: map[string]string{
				// "publicPaths":  "80:/"
				"publicPaths": publicPath,
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				api.Container{
					Name: repoImage.GeneratePodName(),
					//TODO: How would we get default images?
					Image: dockerURI,
					Env: []api.EnvVar{
						api.EnvVar{
							Name:  "PORT",
							Value: port,
						},

						api.EnvVar{
							Name: "PRIVATE_API_KEY",
							ValueFrom: &api.EnvVarSource{
								SecretKeyRef: &api.SecretKeySelector{
									LocalObjectReference: api.LocalObjectReference{
										Name: "routing-keys",
									},

									Key: "private-api-key",
								},
							},
						},

						api.EnvVar{
							Name: "PUBLIC_API_KEY",
							ValueFrom: &api.EnvVarSource{
								SecretKeyRef: &api.SecretKeySelector{
									LocalObjectReference: api.LocalObjectReference{
										Name: "routing-keys",
									},

									Key: "public-api-key",
								},
							},
						},
					},
					Ports: []api.ContainerPort{
						api.ContainerPort{
							ContainerPort: intPort,
						},
					},
				},
			},
		},
	}

	//this pod template spec requires this volume in the RC/deployment to function correctly
	//TODO this seems like an impedence mismatch. Should deployment be adding this to the spec?

	// Volumes: []api.Volume{
	// 		api.Volume{
	// 			Name: "routing-keys",
	// 			Secret: &api.SecretVolumeSource{
	// 				SecretName: "routing",
	// 			},
	// 		},
	// 	},

	json, err := json.Marshal(podTemplate)

	if err != nil {
		return "", err
	}

	return string(json), nil
}
