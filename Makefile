GOCMD=go

GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean
GOFMT=${GOCMD} fmt
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINST=$(GOCMD) install

PACKAGE_FILES=config_sample.yml # frontend

# On Windows, build with Git Bash.
CP=cp
RMR=rm -fr


ifndef CIRCLE_BRANCH
override CIRCLE_BRANCH = latest
else
override CIRCLE_BRANCH = $(shell git rev-parse --abbrev-ref HEAD | sed 's/[^a-zA-Z0-9]/-/g')
endif

all: clean test package

clean:
	${GOCLEAN} ./...
	${RMR} build

fmt:
	${GOFMT} ./...

dep:
	${GOINST} github.com/wailsapp/wails/v2/cmd/wails@latest

test:
	$(GOTEST) -v ./...

build: dep
	wails build

install:
	${GOINST} ./...

ifeq ($(OS),Windows_NT)
package: build
	${CP} build/bin/gogallery.exe gogallery-windows-amd64.exe
	7z a -t7z gogallery-windows-amd64.7z gogallery-windows-amd64.exe ${PACKAGE_FILES}
	${RMR} gogallery-windows-amd64.exe
else
package: build
	${CP} build/bin/gogallery gogallery-linux-amd64
	tar -zcvf gogallery-linux-amd64.tgz gogallery-linux-amd64 ${PACKAGE_FILES}
	${RMR} gogallery-linux-amd64
endif

docker:
	docker build . -t robrotheram/gogallery:$(CIRCLE_BRANCH)
	docker build . -t robrotheram/gogallery:latest

docker-publish:
	docker push robrotheram/gogallery:$(CIRCLE_BRANCH)
	docker push robrotheram/gogallery:latest

update: 
	${GOGET} -u
	${GOMOD} tidy
