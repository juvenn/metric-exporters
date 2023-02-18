package emitters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	exporters "github.com/juvenn/metric-exporters"
	"github.com/rcrowley/go-metrics"
)

// Periodically report to io writer
func NewIOReporter(writer io.Writer, reg metrics.Registry, pollInterval time.Duration) *exporters.Reporter {
	return exporters.NewReporter(reg, pollInterval).WithEmitter(NewIOEmitter(writer))
}

// Emit metrics to file
func NewFileEmitter(path string) (*fileEmitter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewIOEmitter(file), nil
}

// Emit metrics to io writer
func NewIOEmitter(writer io.Writer) *fileEmitter {
	return &fileEmitter{
		writer: writer,
	}
}

// Emit metrics to stdout
func NewStdoutEmitter() *fileEmitter {
	return NewIOEmitter(os.Stdout)
}

// Emit metrics to file as json lines.
type fileEmitter struct {
	writer io.Writer
}

func (this *fileEmitter) Name() string {
	// TODO filename?
	return "file"
}

func (this *fileEmitter) Emit(metrics ...*exporters.Metric) error {
	writer := this.writer
	for _, metric := range metrics {
		line, err := json.Marshal(metric)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(writer, string(line))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *fileEmitter) Close() error {
	writer, ok := this.writer.(io.Closer)
	if ok {
		return writer.Close()
	} else {
		return nil
	}
}
