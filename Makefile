SHELL=/usr/bin/env bash

unexport GOFLAGS

APP=expertdata
GOVERSION:=$(shell go version | cut -d' ' -f 3 | cut -d. -f 2)
ifeq ($(shell expr $(GOVERSION) \< 14), 1)
$(warning Your Golang version is go 1.$(GOVERSION))
$(error Update Golang to version $(shell grep '^go' go.mod))
endif

build: 
	rm -f epik-$(APP)
	go build -o epik-$(APP) ./cmd

install:
	install -C ./epik-$(APP) /usr/local/bin/epik-$(APP)

.PHONY: build