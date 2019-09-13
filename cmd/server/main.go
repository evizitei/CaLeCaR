package main

import (
	"flag"

	"github.com/evizitei/lcr-cache/pkg/cache"
)

func parseArgs() *cache.ServerConf {
	logFile := flag.String("logfile", "./log/server.log", "file to write log outputs to as the server runs")
	dataFile := flag.String("data_file", "./data/test_set_1.json", "file to read working set from")
	flag.Parse()
	return &cache.ServerConf{
		LogFile:  logFile,
		DataFile: dataFile,
	}
}

func main() {
	conf := parseArgs()
	server := cache.NewServer(conf)
	server.Listen()
}
