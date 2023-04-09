package model

import (
	"testing"
)

var testFile File = File{
	Repo:      "Repo",
	URL:       "https://youtube.com/video/11.ts",
	Extension: "ts",
	Folder:    "video",
}

func TestGetReplacedPath(t *testing.T) {
	path, err := testFile.getReplacedPath("https://youtube.com/video/11.ts", 10)
	if err != nil {
		t.Errorf("getReplacedPath failed, %s", err.Error())
	}
	if path != "https://youtube.com/video/10.ts" {
		t.Errorf("getReplacedPath failed not valid, %s", path)
	}
}

func TestGetReplacedFileName(t *testing.T) {
	fileName := testFile.getReplacedFileName("11", 1)
	if fileName != "01" {
		t.Errorf("fileName is not valid")
	}

	fileName = testFile.getReplacedFileName("1", 3)
	if fileName != "3" {
		t.Errorf("fileName is not valid")
	}

	fileName = testFile.getReplacedFileName("test-1", 3)
	if fileName != "test-3" {
		t.Errorf("fileName is not valid")
	}

	fileName = testFile.getReplacedFileName("test_1", 3)
	if fileName != "test_3" {
		t.Errorf("fileName is not valid")
	}

	fileName = testFile.getReplacedFileName("test-1-a1-b1", 3)
	if fileName != "test-3-a1-b1" {
		t.Errorf("fileName is not valid")
	}
}

func TestSetExtension(t *testing.T) {
	fileName := testFile.SetExtension()
	if fileName != "ts" {
		t.Errorf("GetExtension is not valid")
	}
}
