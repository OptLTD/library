module github.com/OptLTD/library/search/driver/redis

go 1.25.0

require (
	github.com/OptLTD/library/search v0.0.0
	github.com/alovn/go-bloomfilter v1.1.0
	github.com/duke-git/lancet/v2 v2.3.7
	github.com/redis/go-redis/v9 v9.16.0
)

require (
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/OptLTD/library/search => ../..
