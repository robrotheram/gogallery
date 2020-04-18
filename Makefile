GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=../gogallery
BINARY_UNIX=$(BINARY_NAME)_unix
CIRCLE_BRANCH:=latest


all: clean test build

dep:
	go get -u github.com/gobuffalo/packr/v2/packr2

test:
	cd server && $(GOTEST) -v ./...

build: build-dashboard build-ui build-server

build-dashboard:
	cd client/dashboard && npm install
	cd client/dashboard && npm run build
	mkdir -p server/ui/dashboard
	cp -r client/dashboard/build/* server/ui/dashboard/.

build-ui:
	cd client/frontend && npm install
	cd client/frontend && npm run build
	mkdir -p server/ui/frontend
	cp -r client/frontend/build/* server/ui/frontend/.

build-server:
	cd server && /go/bin/packr2
	cd server && $(GOBUILD) -o $(BINARY_NAME) -v
	cd server && /go/bin/packr2 clean
	
clean: 
	cd server && $(GOCLEAN)
	cd server && rm -f $(BINARY_NAME)
	cd server && rm -f $(BINARY_UNIX)
	cd server && /go/bin/packr2 clean
	rm -rf *.tar.gz
run:
	cd server && $(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

package:
	tar -czvf gogallery-linux-amd64.tgz gogallery config_sample.yml
# Cross compilation
build-linux:
		cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
docker:
		docker build . -t robrotheram/gogallery:$(CIRCLE_BRANCH)
docker-publish:
		docker push robrotheram/gogallery:$(CIRCLE_BRANCH)