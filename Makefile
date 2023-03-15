PRODUCT_NAME = cheetah 
PACKAGE = bin/$(PRODUCT_NAME)
OS = $(shell uname | awk '{print tolower($0)}')
BASE_PATH = $(shell pwd)
MAIN = $(BASE_PATH)/main.go

all:
	go mod init $(PRODUCT_NAME)
	go mod tidy
	go mod vendor

build:
	GOOS=$(OS) go build -o $(PACKAGE) $(MAIN)

clean: $(info cleaning…)
	@rm -rf vendor mock bin
	@rm -rf go.mod go.sum pkg.list