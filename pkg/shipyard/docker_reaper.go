package shipyard

import "time"

//Reap Performs a reap pass.  Any image older than minAge in the local images of the imageCreator will be deleted.  Returns an error if one occurs
func Reap(minAge time.Duration, imageCreator ImageCreator) error {

	images, err := imageCreator.GetLocalImages()

	if err != nil {
		return err
	}

	minAgeTime := time.Now().Add(minAge)

	for _, image := range *images {
		_, exists := image.Labels[TAG_REPO]

		//doesn't have a label from our system don't remove it
		if !exists {
			continue
		}

		createdTime := time.Unix(image.Created, 0)

		//not old enough, skip it
		if createdTime.After(minAgeTime) {
			LogInfo.Printf("Skipping image %s, it was created before min time of %v. Created time is %v", image.ID, minAgeTime, createdTime)
			continue
		}

		LogInfo.Printf("Removing image with id %s and created time of %v", image.ID, createdTime)

		err := imageCreator.DeleteImageRevisionLocal(image.ID)

		if err != nil {
			LogError.Printf("Unable to remove image, error is %s.  Continuing", err)
		}
	}

	return nil

}

//ReapForever The same as Reap, just wrapped in a loop that runs forever.  This is a convenience method only
func ReapForever(minAge time.Duration, imageCreator ImageCreator, interval time.Duration) {

	for {

		//allocate and wait for the wake message
		wakeTimer := time.NewTimer(interval)

		<-wakeTimer.C

		err := Reap(minAge, imageCreator)

		if err != nil {
			LogError.Printf("Unable to reap images, error is %s", err)
		}

	}
}
