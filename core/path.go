package core

import (
	"log"
	"os"
)

func isdir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func CheckDownloadDir() {
	curdir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	if !isdir(curdir + "/download") {
		os.Mkdir(curdir+"/download", 0755)
	}
}
