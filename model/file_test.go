package model

import "testing"

func MakeTestFile() *File {
	return &File {
		Repo: "Repo",
		FileName: "https://youtube.com/video/11.ts",
		Input: &Input{
			URL: ,
		},
	}
}

func TestgetReplacedFileName(t *testing.T) {

}