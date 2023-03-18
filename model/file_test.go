package model

import (
	"testing"
)

var testFile *File = &File{
	Repo:      "Repo",
	URL:       "https://youtube.com/video/11.ts",
	Separator: "11",
	Folder:    "video",
}

func TestGetReplacedFileName(t *testing.T) {
	fileName := testFile.getReplacedFileName("11.ts", 10)
	if fileName != "10.ts" {
		t.Errorf("fileName is not valid")
	}
}
