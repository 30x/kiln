package shipyard

import (
	"io"
	"os"

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
