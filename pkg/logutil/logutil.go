package logutil

import "log"

func Printf(format string, a ...interface{}) {
	log.Printf(format, a...)
}
