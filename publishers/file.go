package publishers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	exporters "github.com/juvenn/metric-exporters"
)

// File publisher write metrics to file as json lines.
type filePublisher struct {
	writer   io.Writer
	isStdout bool
}

func (pub *filePublisher) Publish(metrics ...exporters.Datum) error {
	writer := pub.writer
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

func (pub *filePublisher) Close() error {
	closer, ok := pub.writer.(io.Closer)
	if ok {
		return closer.Close()
	} else {
		return nil
	}
}

func NewFilePublisher(path string) (*filePublisher, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &filePublisher{
		writer: file,
	}, nil
}

func NewStdoutPublisher() (*filePublisher, error) {
	return &filePublisher{
		writer:   os.Stdout,
		isStdout: true,
	}, nil
}

func NewPipePubliser(pipe *io.PipeWriter) (*filePublisher, error) {
	return &filePublisher{
		writer: pipe,
	}, nil
}
