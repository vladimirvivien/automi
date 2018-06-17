package util

import "github.com/go-faces/logger"

func Log(log logger.Interface, msg ...interface{}) {
	if log == nil {
		return
	}
	log.Print(msg...)
}

func Logf(log logger.Interface, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Printf(format, v...)
}
