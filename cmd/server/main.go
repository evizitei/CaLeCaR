package main

import (
	"flag"

	"github.com/evizitei/lcr-cache/pkg/cache"
)

func parseArgs() *cache.ServerConf {
	logFile := flag.String("logfile", "./log/server.log", "file to write log outputs to as the server runs")
	dataFile := flag.String("data_file", "./data/test_set_1.csv", "file to read working set from")
	cacheType := flag.String("cache_type", "FIFO", "One of (NONE, FIFO, LRU, LFU, LCR, LECAR, LECARAC)")
	cacheSize := flag.Int("cache_size", 1000, "number of entries the cache is able to hold")
	flag.Parse()
	return &cache.ServerConf{
		LogFile:   logFile,
		DataFile:  dataFile,
		CacheType: cacheType,
		CacheSize: *cacheSize,
	}
}

func main() {
	conf := parseArgs()
	server := cache.NewServer(conf)
	server.Listen()
}
