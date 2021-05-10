.POSIX:
.SUFFIXES:

.PHONY: default
default: build

.PHONY: build
build:
	GOOS=linux GOARCH=arm GOARM=6 go build -o seahorse main.go

.PHONY: clean
clean:
	rm -fr seahorse
