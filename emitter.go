package exporters

type Emitter interface {
	Emit(metrics ...any) error
	Close() error
}
