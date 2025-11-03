GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GORUN=$(GOCMD) run
GOMOD=$(GOCMD) mod

BINARY_NAME=shortener
MAIN_PACKAGE=./cmd/app

.PHONY: all build run clean deps tools

all: build

build:
	@echo "Building the application..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PACKAGE)

run:
	@echo "Running the application..."
	$(GORUN) $(MAIN_PACKAGE)

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

tools:
	go get github.com/ilyakaznacheev/cleanenv
	go get github.com/go-playground/validator
	go get github.com/mattn/go-sqlite3
