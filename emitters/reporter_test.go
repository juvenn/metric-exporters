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
	data := make([]exporters.Datum, 0, 8)
	for scanner.Scan() {
		bstr := scanner.Bytes()
		datum := &exporters.Datum{}
		err := json.Unmarshal(bstr, datum)
		if err != nil {
			t.Fatalf("%#v\n", err)
		}
		data = append(data, *datum)
	}
	if len(data) != 2 {
		t.Fatalf("Should report 2 times with last graceful report but got %d", len(data))
	}
	for _, datum := range data {
		if datum.Name != "req" {
			t.Errorf("Name req != %s\n", datum.Name)
		}
		if datum.Type != exporters.TypeCounter {
			t.Errorf("Type %s != %s\n", exporters.TypeCounter, datum.Name)
		}
		if host := datum.Labels["host"]; host != "localhost" {
			t.Errorf("Labels.host localhost != %s\n", host)
		}
	}
	if count := data[1].Fields["count"]; count != 0 {
		t.Errorf("Counter should be reset to 0, but got %f\n", count)
	}
}
