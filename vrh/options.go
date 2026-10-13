package vrh

// Backend selects the scoring backend used by the ring.
type Backend uint8

const (
	// BackendAuto selects the N3 scalar owner: native amd64 for larger rings,
	// portable Go elsewhere.
	BackendAuto Backend = iota

	// BackendScalar selects N3 scalar (native amd64 for larger rings, Go elsewhere).
	BackendScalar

	// BackendAVX2 is reserved for a future N3 implementation. New returns an
	// error for this option, even on CPUs with AVX2.
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
