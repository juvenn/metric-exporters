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
	reshape    Reshape           // metric transformer
	logf       func(format string, a ...any)
}

// Add more emitter to the reporter. Repeatedly apply it to add multiple emitters.
func (rep *Reporter) WithEmitter(emitter Emitter) *Reporter {
	rep.emitters = append(rep.emitters, emitter)
	return rep
}

// Apply a function to transform metric name, labels, fields before emitting.
func (rep *Reporter) WithReshape(fn Reshape) *Reporter {
	rep.reshape = fn
	return rep
}

// Add a global label to each metric. Repeatedly apply it to add multiple labels.
func (rep *Reporter) WithLabel(k, v string) *Reporter {
	if rep.labels == nil {
		rep.labels = make(map[string]string)
	}
	rep.labels[k] = v
	return rep
}

// Auto remove (or not) metric from registry after polled. NOTE that all metrics
// must be dynamically registered to registry via `GetOrRegister`, otherwise
// they will be lost after polled.
func (rep *Reporter) WithAutoRemove(b bool) *Reporter {
	rep.autoRemove = b
	return rep
}

// Start and return reporter, the reporter should be Closed when shutting down.
func (rep *Reporter) Start() (*Reporter, error) {
	rep.exit = make(chan struct{})
	if len(rep.emitters) < 1 {
		return nil, fmt.Errorf("Please specify at least one emitter to report metrics.")
	}
	go rep.loopPoll()
	return rep, nil
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

// Log to customized logger, default to log.Printf.
func (rep *Reporter) WithLogger(fn func(format string, a ...any)) *Reporter {
	rep.logf = fn
	return rep
}

// Poll metrics from registry
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
			if rep.reshape != nil {
				metric = rep.reshape(metric)
			}
			// do not emit if metric has zero fields
			if len(metric.Fields) > 0 {
				points = append(points, metric)
			}
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

func NewReporter(registry metrics.Registry, pollInterval time.Duration) *Reporter {
	rep := &Reporter{
		registry: registry,
		interval: pollInterval,
		logf:     log.Printf,
	}
	return rep
}
