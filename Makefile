# ==============================================================================
# define global Makefile variables for later reference

COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
# project root directory
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR)/ && pwd -P))
# directory for storing build output and temporary files
OUTPUT_DIR := $(ROOT_DIR)/_output

# ==============================================================================
# define the Makefile "all" phony target, which is executed by default when running 'make'
.PHONY: all
all: add-copyright format build

# ==============================================================================
# define other phony targets

.PHONY: add-copyright
add-copyright: # add license header
	addlicense -v -f $(ROOT_DIR)/scripts/licenseHead.txt $(ROOT_DIR) --skip-dirs=third_party,vendor,$(OUTPUT_DIR)

.PHONY: build
build: tidy # compile source code, auto adding/removing dependency packages depending on "tidy" target
	go build -gcflags "all=-N -l" -v -ldflags "$(GO_LDFLAGS)" -o $(OUTPUT_DIR)/dCache $(ROOT_DIR)/main.go

.PHONY: format
format: # format source code
	gofmt -s -w ./

.PHONY: run
run: # run the program
	sh $(ROOT_DIR)/scripts/run.sh

.PHONY: test
test: # run unit tests
	sh $(ROOT_DIR)/scripts/run.sh

.PHONY: tidy
tidy: # auto add/remove dependency packages
	go mod tidy

.PHONY: clean
clean: # clean build output and temporary files
	-rm -vrf $(OUTPUT_DIR)