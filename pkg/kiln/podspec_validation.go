package kiln

import (
	"encoding/json"
	"strings"

	"os"

	"k8s.io/kubernetes/pkg/api"
)

const (
	routableLabelNameDefault         = "routable"
	publicPathsAnnotationNameDefault = "publicPaths"
)

// ValidatePTS validates that the given podspec conforms to Kiln standards
func ValidatePTS(podspec string) (bool, string, error) {
	var routableLabelName, publicPathsAnnotationName string

	pts := api.PodTemplateSpec{}
	LogInfo.Printf("Validating pod template spec\n")

	err := json.Unmarshal([]byte(podspec), &pts)
	if err != nil {
		return false, "failed to Unmarshal as Pod Template Spec", err
	}

	if routableLabelName = os.Getenv("ROUTING_LABEL_SELECTOR"); routableLabelName == "" {
		routableLabelName = routableLabelNameDefault
	}

	if publicPathsAnnotationName = os.Getenv("PATHS_ANNOTATION"); publicPathsAnnotationName == "" {
		publicPathsAnnotationName = publicPathsAnnotationNameDefault
	}

	// validate that PTS contains routable label
	if _, exists := pts.Labels[routableLabelName]; !exists {
		return false, "missing routable label", nil
	}

	// valudate that PTS contains publicPaths annotation
	if _, exists := pts.Annotations[publicPathsAnnotationName]; !exists {
		return false, "missing publicPaths annotation", nil
	}

	// verify that any containers reference images built by Kiln
	dockerRegistry := os.Getenv("DOCKER_REGISTRY_URL")
	for _, container := range pts.Spec.Containers {
		parts := strings.Split(container.Image, "/")

		if len(parts) < 3 {
			return false, "invalid image URI", nil
		} else if parts[0] != dockerRegistry {
			return false, "invalid image URI", nil
		}
	}

	LogInfo.Printf("Pod template spec passes validation\n")
	return true, "", nil
}
