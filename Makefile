PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' \
	-not -path './vendor/*')

${APP_NAME}: $(SOURCES)
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

RELEASE_SOURCES := $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_github_release

.PHONY: clean
clean: clean_common
