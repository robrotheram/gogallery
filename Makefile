GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
FYNE=fyne
BINARY_NAME=gogallery

ifndef CIRCLE_BRANCH
override CIRCLE_BRANCH = latest
else 
override CIRCLE_BRANCH = $(shell git rev-parse --abbrev-ref HEAD | sed 's/[^a-zA-Z0-9]/-/g')
endif

all: clean test build

deps:
	go install fyne.io/fyne/v2/cmd/fyne@latest

build:
	CGO_ENABLED=1 $(FYNE) package -os linux .

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME).app

package:
	tar -czvf gogallery-linux-amd64.tgz $(BINARY_NAME) config_sample.yml themes

# Cross compilation
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(FYNE) package -os linux -name $(BINARY_NAME) .

build-linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc $(FYNE) package -os linux -name $(BINARY_NAME)-arm64 .

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(FYNE) package -os windows -name $(BINARY_NAME).exe .

build-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(FYNE) package -os darwin -name $(BINARY_NAME).app .

build-darwin-arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(FYNE) package -os darwin -name $(BINARY_NAME)-arm64.app .

build-all: build-linux build-linux-arm64 build-windows build-darwin build-darwin-arm64

docker:
		docker build . -t robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker build . -t robrotheram/gogallery:latest
docker-publish:
		docker push robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker push robrotheram/gogallery:latest

install:
	cp $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)

run:
	CGO_ENABLED=1 $(GOCMD) run .

dev:
	$(GOCMD) run . --config config_sample.yml

update: 
	go get -u
	go mod tidy

# Development helpers
fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

lint:
	golangci-lint run

.PHONY: all deps build test clean package build-linux build-linux-arm64 build-windows build-darwin build-darwin-arm64 build-all docker docker-publish install run dev update fmt vet lint