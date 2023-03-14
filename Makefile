PACKAGE = bin/cheetah
OS = $(shell uname | awk '{print tolower($0)}')
BASE_PATH = $(shell pwd)
MAIN = $(BASE_PATH)/main.go

start:
	go mod init $(PACKAGE)
	go mod tidy
	go mod vendor

build:
	GOOS=$(OS) go build -o $(PACKAGE) $(MAIN)
