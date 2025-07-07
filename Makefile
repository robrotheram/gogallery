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

build-themes:
	@echo "Building themes..."
	@for theme in themes/*; do \
		if [ -d "$$theme" ]; then \
			echo "Building theme: $$theme"; \
			cd "$$theme" && npm install && npm run build && npm run clean;  \
			cd -; \
		fi; \
	done

build:
	@echo "Building GoGallery..."
	$(GOBUILD) -o $(BINARY_NAME) -v .