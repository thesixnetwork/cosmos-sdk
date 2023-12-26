#!/usr/bin/make -f


all: cosmovisor test

cosmovisor:
	go build -mod=readonly ./cmd/cosmovisor 

cosmovisor_linux:
	GOOS=linux go build -mod=readonly -o linux_bin/cosmovisor ./cmd/cosmovisor

test:
	go test -mod=readonly -race ./...

.PHONY: all cosmovisor test
