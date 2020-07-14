PWD := $(shell pwd 2>/dev/null)
PKG_LIST := $(shell go list ./... 2>/dev/null)

.PHONY: gofmt
gofmt:
	go fmt $(PKG_LIST)

.PHONY: govet
govet:
	go vet $(PKG_LIST)

.PHONY: gotest
gotest:
	go test -race -short $(PKG_LIST)

.PHONY: examples
examples:
	  sh ./build.sh examples

.PHONY: server
server:
	  sh ./build.sh server

.PHONY: coverage
coverage:
	  rm -rf ./testdata/coverage/*
	  go test -coverpkg=./... -coverprofile=./testdata/coverage/coverage.out ./...
	  go tool cover -html=./testdata/coverage/coverage.out -o ./testdata/coverage/coverage.html
	  go tool cover -func=./testdata/coverage/coverage.out -o ./testdata/coverage/coverage.txt

.PHONY: proto
proto:
ifeq (, $(shell which protoc 2>/dev/null))
	$(error "No protoc in $(PATH)")
endif
	protoc -I proto --go_out=plugins=grpc:. proto/*.proto
