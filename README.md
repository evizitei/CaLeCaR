# lcr-cache
C.S. capstone project, evaluate "least costly recomputation" cache replacement policy as a member of a learned weighted policy choice system.

## Usage

The server that runs caching strategies is in cmd/server.
It can be compiled with `make build` and executed then
with `./bin/server`

This command accepts arguments for changing it's
default behavior:

```bash
./bin/server \
  -logfile ./log/server.log \
  -data_file ./data/test_set_1.csv \
  -cache_type NONE \
  -cache_size 0
```

Set a different cache type by changing the cache_type argument:

```bash
./bin/server \
  -logfile ./log/server.log \
  -data_file ./data/test_set_1.csv \
  -cache_type LECAR \
  -cache_size 20
```

There's a make task for launching this:  `make serve`

One easy way to test the server is to use something like
"nc" (netcat) to poke at the server and fetch values:

```bash
evizitei-ltemp:~ evizitei$ nc localhost 1234
fetch,key1
VALUE:val1
COST:2
```

To try a bunch of queries in order to really exercise the caching
behavior, try using the client program:

```bash
./bin/client \
  -keyfile ./data/client/traffic_set_baseline.csv
```

and you can choose which traffic pattern to use with the
keyfile argument:

```bash
./bin/client -keyfile ./data/client/traffic_set_lru.csv
```

Again, there's a make task: `make query`


### TODO

[-] create http server that logs to file
[-] define cache fetch operation with null response
[-] curl server, get cache results from data reader if not in cache
[-] create tcp client program to fetch a set of keys from server
[-] parameterize server with data file (key, val, cost)
[-] tcp server on start creates data reader
[-] make client consume a list of keys as things to fetch
[-] client log produces cost assessment for each item it fetches
[-] client program summarizes traffic type and cost

[-] parameterize cache size
[-] implement NONE cache replacement strategy
[-] parameterize cache replacement strategy, implement FIFO
[-] implement LRU
[-] implement LFU
[-] implement LCR
[-] implement LeCaR
[ ] build traffic patterns favorable to each existing cache type
[ ] track regret in server
[ ] allow regret reset in server



[ ] wrap tests around extracted functionality
[ ] parameterize server with cache size (in record/key count)
[ ] tcp server on start creates mem cache with declared size constraint

[ ] parameterize port (1234 by default)