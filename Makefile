default: build

build:
	go build -o ./bin/server ./cmd/server
	go build -o ./bin/client ./cmd/client

clean:
	rm bin/*

serve:
	./bin/server \
	  -logfile ./log/server.log \
	  -data_file ./data/test_set_1.csv \
	  -cache_type LRU \
	  -cache_size 250

query:
	./bin/client -keyfile ./data/client/traffic_set_baseline.csv

.PHONY: clean default build serve test
