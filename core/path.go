package core

import (
	"log"
	"os"
	"path/filepath"
)

func CheckDownloadDir() {
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	downloadDir := filepath.Join(curDir, "download")
	if !isDir(downloadDir) {
		if err := os.Mkdir(downloadDir, 0o755); err != nil {
			log.Fatal(err)
		}
	}
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
