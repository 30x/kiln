package shipyard

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"os"

	"k8s.io/kubernetes/pkg/api"
)

//GenerateShipyardTemplateSpec generate a valid pod template spec given the docker image uri, and return it as a json encoded string
func GenerateShipyardTemplateSpec(dockerURI string) (string, error) {

	//TODO validate we only ever have 1 port in the public paths.  Parse out the port and then set it below.

	repoImage, err := NewRepoImage(dockerURI)

	if err != nil {
		return "", err
	}

	port := os.Getenv("POD_PORT")

	if port == "" {
		port = "9000"
	}

	intPort, err := strconv.Atoi(port)

	if err != nil {
		return "", errors.New("Port must be parsable to an int")
	}

	pullSecretName := os.Getenv("PULL_SECRET_NAME")

	if pullSecretName == "" {
		pullSecretName = "ecr-key"
	}

	//the cdir to allow traffic from.  TODO make this space or comma delimited
	cdir := os.Getenv("POD_CDIR")

	if cdir == "" {
		cdir = "10.1.0.0/16"
	}

	podTemplate := api.PodTemplateSpec{
		ObjectMeta: api.ObjectMeta{
			Labels: map[string]string{
				"runtime":  "shipyard",
				"app":      repoImage.GeneratePodName(),
				"routable": "true",
			},
			Annotations: map[string]string{
				//TODO, only allow from same namespace and ingress
				"projectcalico.org/policy": fmt.Sprintf("allow tcp from cidr 192.168.0.0/16; allow tcp from cidr %s", cdir),
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				api.Container{
					Name: repoImage.GeneratePodName(),
					//TODO: How would we get default images?
					Image:           dockerURI,
					ImagePullPolicy: api.PullAlways,
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
										Name: "routing",
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
										Name: "routing",
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
			ImagePullSecrets: []api.LocalObjectReference{
				api.LocalObjectReference{
					Name: pullSecretName,
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
