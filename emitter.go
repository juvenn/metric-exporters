package exporters

// An emitter receive metric points from reporter, and publish them to database.
type Emitter interface {
	// Transform metric points and publish to database
	Emit(metrics ...*Metric) error

	// A user-friendly name to identify upstream, such as db address.
	Name() string

	Close() error
}
