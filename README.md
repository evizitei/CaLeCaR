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
./bin/client -keyfile ./data/client/generated_lru_keys.csv
./bin/client -keyfile ./data/client/generated_lfu_keys.csv
```

You can also submit multiple keyfiles

Again, there's a make task: `make query`

### Available Datasets

There are 10,000 keys in the "working" dataset.  Cache size for each experiment will be fixed at 250, 2.5% of the
size of the dataset itself.

Each set of keys was generated from a helper script to favor each type of cache.
You have to have numpy installed to use it (`pip install numpy`).

The LRU genrator samples from a normal distribution with a standard deviation 1/5 the size of the cache, and it uses samples from the distribution both to choose which keys to query
and to move the center of the distribution around in a random walk.

```
python ./data/key_generator.py \
  --cache_type=LRU \
  --output_filename=./data/client/generated_lru_keys.csv \
  --output_count=100000 \
  --max_key_val=10000
```

The LFU genrator chooses from a "frequently used" set most of the time, and randomly injects
scans of random (useless to cache) runs large enough to purge LRU caches.

```
python ./data/key_generator.py \
  --cache_type=LFU \
  --output_filename=./data/client/generated_lfu_keys.csv \
  --output_count=100000 \
  --max_key_val=10000
```

The LCR genrator has 2 "frequently used" sets, a low cost one and a high cost one.
It works much like the LFU generator, but half the time it pulls the high cost keys.

```
python ./data/key_generator.py \
  --cache_type=LCR \
  --output_filename=./data/client/generated_lcr_keys.csv \
  --output_count=100000 \
  --max_key_val=10000
```

### COMPARISON DATA:

https://docs.google.com/spreadsheets/d/19LT0O388c1sHTvBwB9Q5zxfHvON7khZYlq-MAQOhn4M/edit#gid=0

For Full Dataset

| DATASET | ALGORITHM |      COST     | COST-hitrate |
------------------------------------------------------
| LRU     | NONE      | 3,011,314,923 | 0.000 |
| LRU     | LRU       |   608,685,877 | 0.798 |
| LRU     | LFU       | 2,780,788,465 | 0.077 |
| LRU     | FIFO      |   647,607,222 | 0.785 |
| LRU     | LCR       | 1,292,405,204 | 0.571 |
| LRU     | LECAR     |   669,338,196 | 0.778 |
| LRU     | CALECAR   |   653,562,344 | 0.783 |
---------------------------------------
| LFU     | NONE      | 3,581,836,682 | 0.000 |
| LFU     | LRU       | 3,108,356,918 | 0.132 |
| LFU     | LFU       | 2,476,908,518 | 0.308 |
| LFU     | FIFO      | 3,121,685,770 | 0.128 |
| LFU     | LCR       | 2,509,968,185 | 0.299 |
| LFU     | LECAR     | 3,077,671,875 | 0.141 |
| LFU     | CALECAR   | 2,998,214,157 | 0.163 |
---------------------------------------
| LCR     | NONE      | 4,057,993,944 | 0.000 |
| LCR     | LRU       |  | |
| LCR     | LFU       |  | |
| LCR     | FIFO      |  | |
| LCR     | LCR       | 2,444,963,263 | 0.397 |
| LCR     | LECAR     | 3,685,009,773 | 0.092 |
| LCR     | CALECAR   | 3,600,191,459 | 0.113 |
---------------------------------------
| all-3   | NONE      | 10,651,145,549 | 0.000 |
| all-3   | LRU       |  7,404,534,146 | 0.305 |
| all-3   | LFU       | 10,335,119,024 | 0.030 |
| all-3   | FIFO      |  | |
| all-3   | LCR       |  | |
| all-3   | LECAR     |  | |
| all-3   | CALECAR   |  | |


For first half of dataset.


### TODO

  -[-] create http server that logs to file
  -[-] define cache fetch operation with null response
  -[-] curl server, get cache results from data reader if not in cache
  -[-] create tcp client program to fetch a set of keys from server
  -[-] parameterize server with data file (key, val, cost)
  -[-] tcp server on start creates data reader
  -[-] make client consume a list of keys as things to fetch
  -[-] client log produces cost assessment for each item it fetches
  -[-] client program summarizes traffic type and cost
  -[-] parameterize cache size
  -[-] implement NONE cache replacement strategy
  -[-] parameterize cache replacement strategy, implement FIFO
  -[-] implement LRU
  -[-] implement LFU
  -[-] implement LCR
  -[-] implement LeCaR
  -[-] parameterize server with cache size (in record/key count)
  -[-] tcp server on start creates mem cache with declared size constraint
  -[ ] build traffic patterns favorable to each existing cache type
  -[ ] track regret in server
  -[ ] run experiments highlighting traffic pattern empirical costs
  -[ ] change regret metric for CaLeCar to care about cost
  -[ ] allow regret reset in server
  -[ ] wrap tests around extracted functionality
  -[ ] parameterize port (1234 by default)