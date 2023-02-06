package exporters

// An emitter receive metric points from reporter, and publish them to database.
type Emitter interface {
	// Transform metric points and publish to database
	Emit(metrics ...*Metric) error

	// A user-friendly name to identify upstream, such as db address.
	Name() string

	Close() error
}

// A Reshape is a function that can reshape metric, updating name, labels, or
// fields. An emitter implementation should support WithReshape(fns ...Reshape)
// option, and applying it to transform metric before submitting to upstream.
//
// The most common use case is when labels are often encoded in metric name,
// where a reshape can decode that into labels.
//
//    req.appId.xxx.method.GET: 1 => req appId=xxx,method=GET 1
//
// Reshape should not mutate input metric.
type Reshape func(Metric) Metric
