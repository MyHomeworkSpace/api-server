package api

import (
	"log"
	"runtime"
)

func ErrorLog_LogError(desc string, err error) {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, false)
	stackTrace := string(buf[0:stackSize])

	log.Println("======================================")

	log.Printf("Error occurred while '%s'!", desc)
	errDesc := ""
	if err != nil {
		errDesc = err.Error()
	} else {
		errDesc = "(err == nil)"
	}
	log.Println(errDesc)
	log.Println(stackTrace)

	log.Println("======================================")
}
