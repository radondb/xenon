PREFIX    :=/usr/local
export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:$(PATH)

build: LDFLAGS   += $(shell GOPATH=${GOPATH} src/build/ldflags.sh)
build:
	@echo "--> Building..."
	@mkdir -p bin/
	go build -v -o bin/xenon    --ldflags '$(LDFLAGS)' src/xenon/xenon.go
	go build -v -o bin/xenoncli --ldflags '$(LDFLAGS)' src/cli/cli.go
	@chmod 755 bin/*

clean:
	@echo "--> Cleaning..."
	@mkdir -p bin/
	@go clean
	@rm -f bin/*
	@rm -f coverage*

install:
	@echo "--> Installing..."
	@install bin/xenon bin/xenonctl $(PREFIX)/sbin/

fmt:
	go fmt ./...

test:
	@echo "--> Testing..."
	@$(MAKE) testcommon
	@$(MAKE) testlog
	@$(MAKE) testrpc
	@$(MAKE) testconfig
	@$(MAKE) testmysql
	@$(MAKE) testmysqld
	@$(MAKE) testserver
	@$(MAKE) testraft
	@$(MAKE) testcli

testcommon:
	go test -v xbase/common
testlog:
	go test -v xbase/xlog
testrpc:
	go test -v xbase/xrpc
testconfig:
	go test -v config
testmysql:
	go test -v mysql
testmysqld:
	go test -v mysqld
testserver:
	go test -v server
testraft:
	go test -v raft
testcli:
	go test -v cli/cmd

COVPKGS = xbase/common\
		  xbase/xlog\
		  xbase/xrpc\
		  config\
		  mysql\
		  mysqld\
		  raft\
		  server
vet:
	go vet $(COVPKGS)

coverage:
	go build -v -o bin/gotestcover \
	src/vendor/github.com/pierrre/gotestcover/*.go;
	bin/gotestcover -coverprofile=coverage.out -v $(COVPKGS)
	go tool cover -html=coverage.out
