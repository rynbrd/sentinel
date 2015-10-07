name=sentinel

changed=$(shell git diff --shortstat 2> /dev/null | tail -n1)
version=$(shell git tag --points-at HEAD | tail -n1)
branch=$(shell git rev-parse --abbrev-ref HEAD | tr -c "[[:alnum:]]\\n._-" "_")

export BIN=$(shell pwd)/bin
export GOPATH=$(shell pwd)/.go
org_url=github.com/BlueDragonX
org_path=$(GOPATH)/src/$(org_url)
project_url=$(org_url)/$(name)
project_path=$(org_path)/$(name)

.PHONY: clean test static
all: $(BIN)/$(name)
static: $(BIN)/$(name).static

clean:
	rm -rf $(BIN) $(GOPATH)

$(project_path):
	mkdir -p $(org_path)
	ln -sf $(shell pwd) $(project_path)

$(BIN)/$(name): $(project_path)
	go get -d $(project_url)
	go install -a $(project_url)
	mkdir -p $(BIN)
	mv $(GOPATH)/bin/$(name) $(BIN)/$(name)
$(BIN)/$(name).static: $(project_path)
	go get -d $(project_url)
	go install -a -ldflags "-linkmode external -extldflags -static" $(project_url)
	mkdir -p $(BIN)
	mv $(GOPATH)/bin/$(name) $(BIN)/$(name).static

test: $(project_path)
	test -z "$(shell gofmt -s -l *.go)"
	go vet .
	go get -d -t .
	go test -v -race .
