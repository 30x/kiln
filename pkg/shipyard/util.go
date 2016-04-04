package shipyard

import (
	"io"
	"os"
	"strings"

	"github.com/nu7hatch/gouuid"
)

//CopyFile copy a file from the src to the destination
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	//defer closing
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	//defer closing
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}

func UUIDString() string {
	uuidBinary, err := uuid.NewV4()

	if err != nil {
		return "Ididntgenerateauuid"
	}

	return uuidBinary.String()
}

//GetImageName get the image name from the repository.  If it's nil, none exists
func GetImageName(repositoryName *string) *string {
	parts := strings.Split(*repositoryName, "/")

	//not the right length, drop it
	if len(parts) != 2 {
		return nil
	}

	return &parts[0]
}

//GetImageNameFromTags Get the image names from tags
func GetImageNameFromTags(repositoryNames []string) *string {

	for _, tag := range repositoryNames {

		name := GetImageName(&tag)

		if name != nil {
			return name
		}
	}

	return nil

}
