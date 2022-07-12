.PHONY: build

build:
	go build -o ./bin/dns

run:
	./bin/dns

clean:
	rm -rf ./bin/