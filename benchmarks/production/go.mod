module github.com/dmedovich/VRH/benchmarks/production

go 1.27.1

require (
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
	github.com/dmedovich/VRH v0.0.0
	github.com/nspcc-dev/hrw/v2 v2.0.4
)

require (
	github.com/twmb/murmur3 v1.1.8 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
)

replace github.com/dmedovich/VRH => ../..
