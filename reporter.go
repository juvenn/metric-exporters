package exporters

import (
	"fmt"
	"log"
	"time"

	"github.com/rcrowley/go-metrics"
)

// A reporter periodically cut metrics and publish to given publishers.
type Reporter struct {
	registry   metrics.Registry
	interval   time.Duration // poll and report interval
	autoRemove bool          // auto remove metric such as counter
	emitters   []Emitter
	exit       chan struct{}     // signal when shutting down
	labels     map[string]string // global labels attach to each metric
	logf       func(format string, a ...any)
}

func (rep *Reporter) pollMetrics() []*Metric {
	points := make([]*Metric, 0, 128)
	rep.registry.Each(func(name string, metrik any) {
		metric := CollectMetric(name, metrik)
		if rep.autoRemove {
			// remove metric to keep zero metrics from hanging all time
			rep.registry.Unregister(name)
		}
		if metric != nil {
			if metric.Labels == nil && len(rep.labels) > 0 {
				metric.Labels = make(map[string]string)
			}
			for k, v := range rep.labels {
				metric.Labels[k] = v
			}
			points = append(points, metric)
		}
	})
	return points
}

func (rep *Reporter) loopPoll() {
	rep.logf("Start reporting metrics (every %s) to %s ...", rep.interval, rep.emitters[0].Name())
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
	for _, em := range rep.emitters {
		if err := em.Emit(metrics...); err != nil {
			rep.logf("ERROR: Report %d metric points to %s error: %s\n", len(metrics), em.Name(), err.Error())
		} else {
			rep.logf("Reported %d metric points to %s\n", len(metrics), em.Name())
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
		return nil, fmt.Errorf("Please specify at least one emitter to report metrics.")
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

// Auto remove metric after each report, such as Counter.
func WithAutoRemove(flag bool) Option {
	return func(rep *Reporter) {
		rep.autoRemove = flag
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
