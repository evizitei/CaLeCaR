package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/JohnCGriffin/overflow"
)

type clientConf struct {
	host    string
	keyfile *string
	port    int
}

type queryResult struct {
	value string
	cost  int
}

func parseArgs() *clientConf {
	keyFile := flag.String("keyfile", "./data/client/traffic_set_baseline.csv", "file with series of keys to fetch")
	flag.Parse()
	return &clientConf{
		keyfile: keyFile,
		port:    1234,
		host:    "localhost",
	}
}

func queryKey(conf *clientConf, key string) queryResult {
	conn, err := net.Dial("tcp", conf.host+":"+strconv.Itoa(conf.port))
	if err != nil {
		fmt.Println("ERROR dialing server: ", err)
		os.Exit(-1)
	}
	conn.Write([]byte("fetch," + key))
	connBuff := bufio.NewReader(conn)
	result := queryResult{}
	for {
		response, err := connBuff.ReadString('\n')
		response = strings.Replace(response, "\n", "", -1)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR parsing connection: ", err)
			os.Exit(-1)
		}
		if strings.Contains(response, "VALUE:") {
			responseParts := strings.Split(response, ":")
			result.value = responseParts[1]
		} else if strings.Contains(response, "COST:") {
			responseParts := strings.Split(response, ":")
			cost, err := strconv.Atoi(responseParts[1])
			if err != nil {
				fmt.Println("ERROR parsing cost: ", err, response)
				os.Exit(-1)
			}
			result.cost = cost
		} else {
			fmt.Println("Unsure how to parse response line: ", response)
		}
	}
	conn.Close()
	return result
}

func queryTrafficPattern(conf *clientConf) {
	accumulatedCost := 0
	keysF, err := os.OpenFile(*conf.keyfile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("ERROR reading keyfile: ", err)
		os.Exit(-1)
	}
	reader := csv.NewReader(keysF)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR reading row of keyfile: ", err)
			os.Exit(-1)
		}
		key := row[0]
		result := queryKey(conf, key)
		fmt.Println("QUERY RESULT: key->" + key +
			", val->" + result.value +
			", cost->" + strconv.Itoa(result.cost))
		accumulatedCost = overflow.Addp(accumulatedCost, result.cost)
	}
	fmt.Println("TRAFFIC COST: ", accumulatedCost)
}

func main() {
	conf := parseArgs()
	queryTrafficPattern(conf)
	fmt.Println("Done!")
}
