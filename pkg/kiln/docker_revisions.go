package kiln

import "strconv"

const defaultRevision = "1"

// AutoRevision Determines the revision tag of an image based on the latest revision
func AutoRevision(repoName string, application string, imageCreator ImageCreator) (string, error) {
	// retrieve all revisions of this image
	images, err := imageCreator.GetImages(repoName, application)

	if err != nil {
		return "", err
	}

	length := len(*images)

	// first of its kind, give it default revision
	if length == 0 {
		return defaultRevision, err
	}

	// just increment to next revision number
	length++

	return strconv.Itoa(length), nil
}
