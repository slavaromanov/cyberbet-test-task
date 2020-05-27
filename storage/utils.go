package storage

import (
	"os"
)

func fileExists(fileName string) bool {
	var info, err = os.Stat(fileName)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
