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

# Cross-platform Fyne build (mimics CI logic)
.PHONY: fyne-build
fyne-build:
	@echo "Building GoGallery with Fyne packaging..."
	go mod download
	@export GOOS=$(GOOS); export GOARCH=$(GOARCH); export CGO_ENABLED=1; \
	if [ "$(GOOS)" = "windows" ]; then \
		OUTPUT_NAME="$(BINARY_NAME).exe"; \
		$(FYNE) package -os windows -name "$$OUTPUT_NAME" .; \
	elif [ "$(GOOS)" = "darwin" ]; then \
		$(FYNE) package -os darwin -name "$(BINARY_NAME).app" .; \
	else \
		OUTPUT_NAME="$(BINARY_NAME)"; \
		$(FYNE) package -os linux -name "$$OUTPUT_NAME" .; \
	fi

# Install Fyne CLI
.PHONY: fyne-cli
fyne-cli:
	$(GOCMD) install fyne.io/tools/cmd/fyne@latest

# Install Fyne dependencies (Ubuntu)
.PHONY: fyne-deps-ubuntu
fyne-deps-ubuntu:
	sudo apt-get update
	sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev

# Install Fyne dependencies (RedHat/Fedora)
.PHONY: fyne-deps-fedora
fyne-deps-fedora:
	sudo dnf install -y gcc mesa-libGL-devel libX11-devel libxkbcommon-devel

# Build all (default)
.PHONY: all
all: fyne-cli fyne-build

# Clean
.PHONY: clean
clean:
	rm -rf $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME).app