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
	npm install -g yarn

test:
	cd server && $(GOTEST) -v ./...

build: build-dashboard build-server

build-dashboard:
	cd client/dashboard && yarn
	cd client/dashboard && yarn run build
	mkdir -p ui/dashboard
	cp -r client/dashboard/build/* ui/dashboard/.

build-server:
	cd server && go generate embeds/ui.go
	cd server && $(GOBUILD) -o $(BINARY_NAME) -v
	
clean: 
	cd server && $(GOCLEAN)
	cd server && rm -f $(BINARY_NAME)
	cd server && rm -f $(BINARY_UNIX)
	rm -rf ui/frontend
	rm -rf ui/dashboard
run:
	cd server && $(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

package:
	tar -czvf gogallery-linux-amd64.tgz gogallery config_sample.yml ui
# Cross compilation
build-linux:
		cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
docker:
		docker build . -t robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker build . -t robrotheram/gogallery:latest
docker-publish:
		docker push robrotheram/gogallery:$(CIRCLE_BRANCH)
		docker push robrotheram/gogallery:latest