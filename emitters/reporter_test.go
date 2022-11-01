package emitters

import (
	"bufio"
	"encoding/json"
	"io"
	"testing"
	"time"

	exporters "github.com/juvenn/metric-exporters"
	"github.com/rcrowley/go-metrics"
)

func TestReportToFile(t *testing.T) {
	reg := metrics.NewRegistry()
	pr, pw := io.Pipe()
	defer pr.Close()
	piper := NewIOEmitter(pw)
	stdout := NewStdoutEmitter()
	rep, err := exporters.NewReporter(reg, 800*time.Millisecond,
		exporters.WithLabels("host", "localhost"),
		exporters.WithAutoReset(true),
		exporters.WithEmitters(piper, stdout))
	if err != nil {
		t.Fatalf("%#v\n", err)
	}
	rep.Start()
	counter := metrics.NewCounter()
	reg.Register("req", counter)
	counter.Inc(1)
	go func() {
		time.Sleep(1 * time.Second)
		rep.Close()
	}()

	scanner := bufio.NewScanner(pr)
	data := make([]exporters.Metric, 0, 8)
	for scanner.Scan() {
		bstr := scanner.Bytes()
		metric := &exporters.Metric{}
		err := json.Unmarshal(bstr, metric)
		if err != nil {
			t.Fatalf("%#v\n", err)
		}
		data = append(data, *metric)
	}
	if len(data) != 2 {
		t.Fatalf("Should report 2 times with last graceful report but got %d", len(data))
	}
	for _, metric := range data {
		if metric.Name != "req" {
			t.Errorf("Name req != %s\n", metric.Name)
		}
		if metric.Type != exporters.TypeCounter {
			t.Errorf("Type %s != %s\n", exporters.TypeCounter, metric.Name)
		}
		if host := metric.Labels["host"]; host != "localhost" {
			t.Errorf("Labels.host localhost != %s\n", host)
		}
	}
	if count := data[1].Fields["count"]; count != 0 {
		t.Errorf("Counter should be reset to 0, but got %f\n", count)
	}
}
