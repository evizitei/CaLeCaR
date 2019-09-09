package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

/*ServerConf holds the cmd flags and other
config params for parameterizing the cache
server*/
type ServerConf struct {
	LogFile *string
}

func parseArgs() *ServerConf {
	logFile := flag.String("logfile", "./log/server.log", "file to write log outputs to as the server runs")
	flag.Parse()
	return &ServerConf{
		LogFile: logFile,
	}
}

func main() {
	conf := parseArgs()
	logFile, err := os.OpenFile(*conf.LogFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("ERROR ", err)
		os.Exit(-1)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)
	logger.Println("Starting cache server...")
	os.Exit(0)
}
