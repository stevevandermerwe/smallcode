.PHONY: build run clean tidy

BINARY=smallcode

build:
	go build -o dist/$(BINARY) .

run: build
	./dist/$(BINARY)

clean:
	rm -rf dist

tidy:
	go mod tidy
