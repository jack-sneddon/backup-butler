BINARY=bin/tracker

all: fmt tidy vet build

fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...

# test:
# 	go test ./...

build:
	go build -o bin/backup-butler ./cmd/backup-butler
