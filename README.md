# lcr-cache
C.S. capstone project, evaluate "least costly recomputation" cache replacement policy as a member of a learned weighted policy choice system.

### TODO

[ ] create http server that logs to file
[ ] define cache fetch operation with null response
[ ] create http client program to fetch a set of keys from server
[ ] parameterize server with data file (key, val, cost)
[ ] http server on start creates data reader
[ ] parameterize server with cache size (in record/key count)
[ ] http server on start creates mem cache with declared size constraint
[ ] curl server, get cache results from data reader if not in cache
[ ] parameterize cache replacement strategy, implement FIFO
[ ] make client consume a list of keys as things to fetch
[ ] client log produces cost assessment for each item it fetches
[ ] client program summarizes traffic type and cost