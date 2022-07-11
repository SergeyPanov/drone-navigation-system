.PHONY: build

build:
	go build -o dns

run: build
	./dns