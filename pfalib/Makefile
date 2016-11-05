PATH  := "$(PATH):$(GOBIN):$(GOPATH)/bin"
SHELL := env PATH=$(PATH) /bin/bash

all: archive.capnp.go pfalib

archive.capnp.go: archive.capnp
	capnp compile -I${GOPATH}/src/zombiezen.com/go/capnproto2/std -ogo archive.capnp

pfalib:
	go build github.com/holgerBerger/pfa/pfalib

test:
	(cd ..;go test)
