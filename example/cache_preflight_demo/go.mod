module example/cache_preflight_demo

go 1.22

replace github.com/ivikasavnish/goroutine => ../..

require github.com/ivikasavnish/goroutine v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.17.2 // indirect
)
