package exporters

import (
	"fmt"
	"log"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Publisher interface {
	Publish(data ...Datum) error
	Close() error
}

type Emitter interface {
	Emit(metrics ...any) error
	Close() error
}

type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeMeter     MetricType = "meter"
	TypeTimer     MetricType = "timer"
	TypeHistogram MetricType = "histogram"
)

// A datum is one data point from metric
type Datum struct {
	Name   string             `json:"name"`
	Type   MetricType         `json:"type"`
	Time   time.Time          `json:"time"`
	Labels map[string]string  `json:"labels,omitempty"`
	Fields map[string]float64 `json:"fields"`
}

func DatumFromMetric(name string, metric any, reset bool) *Datum {
	now := time.Now()
	switch metric := metric.(type) {
	case metrics.Counter:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"count": float64(ms.Count()),
		}
		if reset {
			metric.Clear()
		}
		return &Datum{Name: name, Type: TypeCounter, Time: now, Fields: fields}
	case metrics.Histogram:
		ms := metric.Snapshot()
		ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
		fields := map[string]float64{
			"count":    float64(ms.Count()),
			"max":      float64(ms.Max()),
			"mean":     ms.Mean(),
			"min":      float64(ms.Min()),
			"stddev":   ms.StdDev(),
			"variance": ms.Variance(),
			"p50":      ps[0],
			"p75":      ps[1],
			"p95":      ps[2],
			"p99":      ps[3],
			"p999":     ps[4],
			"p9999":    ps[5],
		}
		if reset {
			metric.Clear()
		}
		return &Datum{Name: name, Type: TypeHistogram, Time: now, Fields: fields}
	case metrics.Meter:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"count": float64(ms.Count()),
			"m1":    ms.Rate1(),
			"m5":    ms.Rate5(),
			"m15":   ms.Rate15(),
			"mean":  ms.RateMean(),
		}
		return &Datum{Name: name, Type: TypeMeter, Time: now, Fields: fields}
	case metrics.Timer:
		ms := metric.Snapshot()
		ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
		fields := map[string]float64{
			"count":    float64(ms.Count()),
			"max":      float64(ms.Max()),
			"mean":     ms.Mean(),
			"min":      float64(ms.Min()),
			"stddev":   ms.StdDev(),
			"variance": ms.Variance(),
			"p50":      ps[0],
			"p75":      ps[1],
			"p95":      ps[2],
			"p99":      ps[3],
			"p999":     ps[4],
			"p9999":    ps[5],
			"m1":       ms.Rate1(),
			"m5":       ms.Rate5(),
			"m15":      ms.Rate15(),
			"meanrate": ms.RateMean(),
		}
		return &Datum{Name: name, Type: TypeTimer, Time: now, Fields: fields}
	case metrics.Gauge:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"gauge": float64(ms.Value()),
		}
		return &Datum{Name: name, Type: TypeGauge, Time: now, Fields: fields}
	case metrics.GaugeFloat64:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"gauge": ms.Value(),
		}
		return &Datum{Name: name, Type: TypeGauge, Time: now, Fields: fields}
	}
	return nil
}

// A reporter periodically cut metrics and publish to given publishers.
type Reporter struct {
	registry  metrics.Registry
	interval  time.Duration // poll and report interval
	autoReset bool          // auto reset metric such as counter
	emitters  []Emitter
	exit      chan struct{}     // signal when shutting down
	labels    map[string]string // global labels attach to each metric
	logf      func(format string, a ...any)
}

func (rep *Reporter) pollMetrics() []Datum {
	data := make([]Datum, 0, 128)
	rep.registry.Each(func(name string, metric any) {
		datum := DatumFromMetric(name, metric, rep.autoReset)
		if datum != nil {
			if datum.Labels == nil && len(rep.labels) > 0 {
				datum.Labels = make(map[string]string)
			}
			for k, v := range rep.labels {
				datum.Labels[k] = v
			}
			data = append(data, *datum)
		}
	})
	return data
}

func (rep *Reporter) loopPoll() {
	// started to report to ...
	rep.logf("Started to report to by %s\n", rep.interval)
	ticker := time.Tick(rep.interval)
	for {
		select {
		case <-rep.exit:
			return
		case <-ticker:
			rep.report()
		}
	}
}

func (rep *Reporter) report() {
	metrics := rep.pollMetrics()
	if len(metrics) == 0 {
		return
	}
	points := make([]any, 0, len(metrics))
	for _, m := range metrics {
		points = append(points, m)
	}
	for _, em := range rep.emitters {
		if err := em.Emit(points...); err != nil {
			rep.logf("Report %d metric points error %#v\n", len(points), err.Error())
		} else {
			rep.logf("Reported %d metric points\n", len(points))
		}
	}
}

func (rep *Reporter) Start() {
	rep.exit = make(chan struct{})
	go rep.loopPoll()
}

// Close reporter and emitters gracefully
func (rep *Reporter) Close() error {
	close(rep.exit)
	rep.report()
	var err error
	for _, em := range rep.emitters {
		err = em.Close()
	}
	return err
}

// Create reporter that is yet to be started. Upon closing, the associated publisher
// will be closed too.
func NewReporter(registry metrics.Registry, interval time.Duration, opts ...Option) (*Reporter, error) {
	rep := &Reporter{
		registry: registry,
		interval: interval,
		logf:     log.Printf,
	}
	for _, opt := range opts {
		opt(rep)
	}
	if len(rep.emitters) < 1 {
		return nil, fmt.Errorf("Please specify at least one publisher to report metrics to.")
	}
	return rep, nil
}

type Option func(*Reporter)

// Where to emit metrics
func WithEmitters(emitters ...Emitter) Option {
	return func(rep *Reporter) {
		rep.emitters = append(rep.emitters, emitters...)
	}
}

// Set poll and report interval, default to 1min
func WithPollInterval(interval time.Duration) Option {
	return func(rep *Reporter) {
		rep.interval = interval
	}
}

// Auto reset metric after each report, such as Counter.
func WithAutoReset(flag bool) Option {
	return func(rep *Reporter) {
		rep.autoReset = flag
	}
}

// Labels that will be attached to each metric. Args must be in the form
// of k,v,k,v.
func WithLabels(kvs ...string) Option {
	n := len(kvs)
	if n%2 != 0 {
		panic("Reporter labels expects an even number of args.")
	}
	labels := make(map[string]string)
	for i := 0; i < n-1; i = +2 {
		k := kvs[i]
		v := kvs[i+1]
		labels[k] = v
	}
	return func(rep *Reporter) {
		rep.labels = labels
	}
}

// How should we print log, default to log.Printf
func WithLogfn(fn func(format string, a ...any)) Option {
	return func(rep *Reporter) {
		rep.logf = fn
	}
}
