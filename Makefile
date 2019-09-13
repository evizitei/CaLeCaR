default: build

build:
	go build -o ./bin/server ./cmd/server

clean:
	rm bin/*

serve:
	./bin/server -logfile ./log/server.log -data_file ./data/test_set_1.csv

.PHONY: clean default build serve test
