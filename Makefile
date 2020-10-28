all: install

install: go.sum
	go install -mod=readonly

build: go.sum
	go build -mod=readonly -o build/brooce

build-linux: go.sum
	GOOS=linux GOARCH=amd64 $(MAKE) build

go.sum: go.mod
	GO111MODULE=on go mod verify
