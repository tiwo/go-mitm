mitm: ./cmd/mitm/*.go ./*.go
	go build ./cmd/mitm

clean: correct-directory
	rm -f ./mitm
.PHONY: clean

correct-directory:
	@grep -q "8ce4a751-4bf8-4c9a-b256-65f92f17076b" ./Makefile
.PHONY: correct-directory