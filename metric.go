package exporters

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
)

type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeMeter     MetricType = "meter"
	TypeTimer     MetricType = "timer"
	TypeHistogram MetricType = "histogram"
)

type Metric struct {
	Name   string             `json:"name"`
	Type   MetricType         `json:"type"`
	Time   time.Time          `json:"time"`
	Labels map[string]string  `json:"labels,omitempty"`
	Fields map[string]float64 `json:"fields"`
}

func CollectMetric(name string, metric any, reset bool) *Metric {
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
		return &Metric{Name: name, Type: TypeCounter, Time: now, Fields: fields}
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
		return &Metric{Name: name, Type: TypeHistogram, Time: now, Fields: fields}
	case metrics.Meter:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"count": float64(ms.Count()),
			"m1":    ms.Rate1(),
			"m5":    ms.Rate5(),
			"m15":   ms.Rate15(),
			"mean":  ms.RateMean(),
		}
		return &Metric{Name: name, Type: TypeMeter, Time: now, Fields: fields}
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
		return &Metric{Name: name, Type: TypeTimer, Time: now, Fields: fields}
	case metrics.Gauge:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"gauge": float64(ms.Value()),
		}
		return &Metric{Name: name, Type: TypeGauge, Time: now, Fields: fields}
	case metrics.GaugeFloat64:
		ms := metric.Snapshot()
		fields := map[string]float64{
			"gauge": ms.Value(),
		}
		return &Metric{Name: name, Type: TypeGauge, Time: now, Fields: fields}
	}
	return nil
}

// Encode metric to prometheus lines, each field will be appended to name
// to produce a new line. Thus a metric with multiple fields will generate
// multiple lines. Trailing line is omitted. See test for examples.
//
//    name_count{region="us-west-2",host="node1"} 1027 1395066363000
//    name_mean{region="us-west-2",host="node1"} 50 1395066363000
//    name_max{region="us-west-2",host="node1"} 110 1395066363000
func (metric *Metric) EncodePromLines() string {
	var buf strings.Builder
	for _, entry := range sortByKey(metric.Labels) {
		k := entry.Key
		v := entry.Val
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(fmt.Sprintf("%s=%q", k, v))
	}
	labels := buf.String()
	if len(labels) != 0 {
		labels = fmt.Sprintf("{%s}", labels)
	}

	ts := metric.Time.UnixMilli()
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	var lines strings.Builder
	for f, v := range metric.Fields {
		if lines.Len() > 0 {
			lines.WriteString("\n")
		}
		// name_field{method="post",code="200"} 20 1395066363000
		line := fmt.Sprintf("%s_%s%s %g %d", metric.Name, f, labels, v, ts)
		lines.WriteString(line)
	}
	return lines.String()
}

// Encode metric as influx line protocol
func (metric *Metric) EncodeInfluxLine(precision string) string {
	var sb strings.Builder
	sb.WriteString(metric.Name)
	// append labels
	for _, entry := range sortByKey(metric.Labels) {
		k, v := entry.Key, entry.Val
		sb.WriteString(",")
		sb.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	sb.WriteString(" ")
	// write fields
	var i int
	for k, v := range metric.Fields {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%g", k, v))
		i++
	}
	// write timestamp
	ts := metric.Time
	sb.WriteString(" ")
	var tss string
	switch precision {
	case "ns":
		tss = fmt.Sprintf("%d", ts.UnixNano())
	case "u", "us":
		tss = fmt.Sprintf("%d", ts.UnixMicro())
	case "ms":
		tss = fmt.Sprintf("%d", ts.UnixMilli())
	default:
		tss = fmt.Sprintf("%d", ts.Unix())
	}
	sb.WriteString(tss)
	return sb.String()
}

type entry struct {
	Key string
	Val string
}

// Sort map by key
func sortByKey(m map[string]string) []entry {
	pairs := make([]entry, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, entry{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})
	return pairs
}
