package shipyard

import (
	"log"
	"os"
)

//Log the info logger
var Log = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
