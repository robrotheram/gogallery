GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=../gogallery
BINARY_UNIX=$(BINARY_NAME)_unix

all: clean test build

dep:
	npm install -g yarn

test:
	cd server && $(GOTEST) -v ./...

build: build-dashboard build-ui build-server

build-dashboard:
	cd client/dashboard && yarn
	cd client/dashboard && yarn run build
	mkdir -p ui/dashboard
	cp -r client/dashboard/build/* ui/dashboard/.

build-ui:
	cd client/frontend && yarn
	cd client/frontend && yarn run build
	mkdir -p ui/frontend
	cp -r client/frontend/build/* ui/frontend/.

build-server:
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
		docker build . -t robrotheram/gogallery:$(git rev-parse --short HEAD)
		docker build . -t robrotheram/gogallery:latest
docker-publish:
		docker push robrotheram/gogallery:$(git rev-parse --short HEAD)
		docker push robrotheram/gogallery:latest