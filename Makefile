
#How to reuse:
#Change the APP variable with the name of you application (/cmd/APP_NAME.go file)

APP=pitcher

OUTPUT=build

.PHONY: clean all run install
all: clean test lint build

clean:
	@echo -e "\nCLEANING $(OUTPUT) DIRECTORY"
	rm -rf ./$(OUTPUT)

$(OUTPUT)/$(APP): build_folder
	@echo -e "\nBUILDING $(OUTPUT)/$(APP) BINARY"
	CGO_ENABLED=0 go build  -o $(OUTPUT)/$(APP) ./cmd

test: build_folder
	@echo -e "\nTESTING"
	go test -v -coverprofile=$(OUTPUT)/coverage.out ./...

build: clean test $(OUTPUT)/$(APP)

install:
	go install ./cmd

run: build
	./$(OUTPUT)/$(APP)

lint:
	go vet ./...

build_folder:
	mkdir -p $(OUTPUT)