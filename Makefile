.PHONY: build test clean

build:
	go build ./...

test:
	go test ./...

clean:
	rm tools/monstrun/monstrun
