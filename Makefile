ifeq ($(OS),Windows_NT)
  PLATFORM="Windows"
else
  ifeq ($(shell uname),Darwin)
    PLATFORM="MacOS"
  else
    PLATFORM="Linux"
  endif
endif

all: build 
export GO111MODULE=on
export GOFLAGS=-mod=vendor
XUPEROS_ROOT := ${PWD}
export XUPEROS_ROOT

build:
	PLATFORM=$(PLATFORM) ./build.sh

test:
	go test -cover ./...

clean:
	PLATFORM=$(PLATFORM) ./build.sh clean

.PHONY: all test clean
