PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' \
	-not -path './vendor/*')

$(APP_NAME): cmd/main/$(APP_NAME)

cmd/main/$(APP_NAME): $(SOURCES)
	cd cmd/main/ && go build -ldflags '-X main.Version=${VERSION}' -o ../../${APP_NAME}

RELEASE_SOURCES := $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_github_release

.PHONY: clean
clean: clean_common
