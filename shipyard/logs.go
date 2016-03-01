package shipyard

import (
	"log"
	"os"
	"syscall"
)

//the constant writer for standard out
var std_out = os.NewFile(uintptr(syscall.Stdout), "/dev/stdOut")

//LOG_INFO the info logger
var Log = log.New(std_out, "INFO: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
