package exporters

type Emitter interface {
	Emit(metrics ...*Metric) error
	Close() error
}
