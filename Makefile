# init project PATH
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output
XVMDIR  := $(HOMEDIR)/xvm
TESTNETDIR := $(HOMEDIR)/testnet

# init command params
export GO111MODULE=on
X_ROOT_PATH := $(HOMEDIR)
export X_ROOT_PATH
export PATH := $(OUTDIR)/bin:$(XVMDIR):$(PATH)

# make, make all
all: clean compile

# make compile, go build
compile: xchain
xchain:
	bash $(HOMEDIR)/auto/build.sh

# make xvm
xvm:
	bash $(HOMEDIR)/auto/build_xvm.sh

# make test, test your code
test: xvm unit
unit:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

# make clean
cleanall: clean cleantest cleanxvm
clean:
	rm -rf $(OUTDIR)
cleantest:
	rm -rf $(TESTNETDIR)
cleanxvm:
	rm -rf $(XVMDIR)

# deploy test network
testnet:
	bash $(HOMEDIR)/auto/deploy_testnet.sh

# avoid filename conflict and speed up build
.PHONY: all compile test clean
