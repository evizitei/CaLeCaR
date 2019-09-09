default: build

build:
	go build -o ./bin/server ./cmd/server

clean:
	rm bin/*

.PHONY: clean default build test
