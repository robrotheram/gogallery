GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=../gogallery
BINARY_UNIX=$(BINARY_NAME)_unix

ifndef CIRCLE_BRANCH
override CIRCLE_BRANCH = latest
else 
override CIRCLE_BRANCH = $(shell git rev-parse --abbrev-ref HEAD | sed 's/[^a-zA-Z0-9]/-/g')
endif

all: clean test build

dep:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

test:
	cd server && $(GOTEST) -v ./...

build: build-linux

package:
	tar -czvf gogallery-linux-amd64.tgz gogallery config_sample.yml ui
# Cross compilation
build-linux:
		wails build
docker:
		docker build . -t robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker build . -t robrotheram/gogallery:latest
docker-publish:
		docker push robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker push robrotheram/gogallery:latest

install:
	cp gogallery /home/robertfletcher/.local/bin/gogallery