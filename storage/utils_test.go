package storage

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	var file, err = ioutil.TempFile("", "temp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if !fileExists(file.Name()) {
		t.Error("File exists!")
	}
}

func TestFileExistsIsDir(t *testing.T) {
	if fileExists(os.TempDir()) {
		t.Error("File is dir!")
	}
}
