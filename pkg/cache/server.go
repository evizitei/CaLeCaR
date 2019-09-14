package cache

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

/*ServerConf holds the cmd flags and other
config params for parameterizing the cache
server*/
type ServerConf struct {
	LogFile  *string
	DataFile *string
}

type cacheEntry struct {
	value string
	cost  int
}

/*Server is the type that listens for
fetch requests and returns them from the data file*/
type Server struct {
	config  *ServerConf
	dataset *map[string]cacheEntry
	logger  *log.Logger
}

func (s *Server) handleConnection(c net.Conn) {
	buf := make([]byte, 1024)
	_, err := c.Read(buf)
	if err != nil {
		s.logger.Println("Conn error: ", err.Error())
		c.Write([]byte("Read Failure, check logs..."))
		c.Close()
		return
	}
	messageValue := string(bytes.Trim(buf, "\x00"))
	messageParts := strings.Split(messageValue, ",")
	command := messageParts[0]
	if command == "fetch" {
		fetchKey := strings.TrimSpace(strings.Replace(messageParts[1], "\n", "", -1))
		s.logger.Println("Fetching ", fetchKey)
		entry, ok := (*s.dataset)[fetchKey]
		if !ok {
			s.logger.Println("No Entry for |" + fetchKey + "|")
			c.Write([]byte("No Entry For Key: " + fetchKey + "\n"))
		} else {
			c.Write([]byte("VALUE:" + entry.value + "\n"))
			c.Write([]byte("COST:" + strconv.Itoa(entry.cost) + "\n"))
		}
		c.Close()
	} else {
		s.logger.Println("No such command: ", command)
		c.Write([]byte("Bad Command"))
		c.Close()
	}
}

/*Listen is how you kick off a serve
loop to wait for incoing connections*/
func (s *Server) Listen() {
	s.logger.Println("Starting cache server...")
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		s.logger.Fatalln("Could not start server: ", err.Error())
		os.Exit(-1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			s.logger.Println("WARNING: Failed to handle request: ", err.Error())
			continue
		}
		go s.handleConnection(conn)
	}
}

func buildLogger(logfile *string) *log.Logger {
	logFile, err := os.OpenFile(*logfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("ERROR ", err)
		os.Exit(-1)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)
	return logger
}

func loadDataset(datafile *string) *map[string]cacheEntry {
	dataMap := make(map[string]cacheEntry)
	dFile, err := os.OpenFile(*datafile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("ERROR opening dataset: ", err)
		os.Exit(-1)
	}
	reader := csv.NewReader(dFile)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading csv file: ", err)
		}
		cost, err := strconv.Atoi(row[2])
		if err != nil {
			fmt.Println("Error reading cost value from file: ", err)
		}
		entry := cacheEntry{
			value: row[1],
			cost:  cost,
		}
		dataMap[row[0]] = entry
	}
	return &dataMap
}

/*NewServer is a constructor for building a new server
with config onboard */
func NewServer(conf *ServerConf) *Server {
	return &Server{
		config:  conf,
		dataset: loadDataset(conf.DataFile),
		logger:  buildLogger(conf.LogFile),
	}
}
