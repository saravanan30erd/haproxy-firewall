package utils

import (
	"log"
	"os"
)

var (
	Infolog  *log.Logger
	Errorlog *log.Logger
)

func NewLog(logpath string) {
	file, err := os.Create(logpath)
	if err != nil {
		panic(err.Error())
	}
	Infolog = log.New(file, "INFO: ", log.LstdFlags|log.Lshortfile)
	Errorlog = log.New(file, "ERROR: ", log.LstdFlags|log.Lshortfile)
}
