PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))
MAIN_APP_DIR = cmd/main

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' \
	-not -path './vendor/*')

${FULL_APP_PATH}: $(SOURCES)
	cd $(MAIN_APP_DIR) && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}

RELEASE_SOURCES := $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_github_release prepare_githooks

.PHONY: clean
clean: clean_common
