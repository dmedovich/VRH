package vrh

// Backend selects the scoring backend used by the ring.
type Backend uint8

const (
	// BackendAuto selects the best backend available on the current CPU.
	BackendAuto Backend = iota

	// BackendScalar forces the portable scalar implementation.
	BackendScalar

	// BackendAVX2 requests the AVX2 implementation.
	//
	// If AVX2 support is not available, Ring construction will return an error.
	BackendAVX2
)

// Options controls Ring construction.
type Options struct {
	Backend Backend
}

// Option modifies Ring construction options.
type Option func(*Options)

// WithBackend selects the requested scoring backend.
func WithBackend(backend Backend) Option {
	return func(opts *Options) {
		opts.Backend = backend
	}
}

// defaultOptions returns the default configuration.
func defaultOptions() Options {
	return Options{
		Backend: BackendAuto,
	}
}
