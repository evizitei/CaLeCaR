package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
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

func handleConnection(c net.Conn, logger *log.Logger) {
	buf := make([]byte, 1024)
	reqLen, err := c.Read(buf)
	if err != nil {
		logger.Println("Conn error: ", err.Error())
		c.Write([]byte("Read Failure, check logs..."))
		c.Close()
		return
	}
	logger.Println("Message Length: ", reqLen)
	c.Write([]byte("Message got read!"))
	c.Write(buf)
	c.Close()
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
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		logger.Fatalln("Could not start server: ", err.Error())
		os.Exit(-1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Println("WARNING: Failed to handle request: ", err.Error())
			continue
		}
		go handleConnection(conn, logger)
	}
	os.Exit(0)
}
