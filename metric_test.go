package exporters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeInfluxLine(t *testing.T) {
	cases := []struct {
		metric *Metric
		out    string
	}{
		{
			metric: &Metric{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost"},
				Fields: map[string]float64{"count": 1}},
			out: "req,host=localhost count=1 1667123357",
		},
		{
			metric: &Metric{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost", "region": "us-west-2"},
				Fields: map[string]float64{"count": 1, "max": 10}},
			out: "req,host=localhost,region=us-west-2 count=1,max=10 1667123357",
		},
		{
			// name with labels
			metric: &Metric{Name: "req,method=POST", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost", "region": "us-west-2"},
				Fields: map[string]float64{"count": 1, "max": 10}},
			out: "req,method=POST,host=localhost,region=us-west-2 count=1,max=10 1667123357",
		},
	}
	assert := assert.New(t)
	for _, tc := range cases {
		line := tc.metric.EncodeInfluxLine("s")
		assert.Equal(line, tc.out)
	}
}

func TestEncodePromLines(t *testing.T) {
	cases := []struct {
		metric *Metric
		out    string
	}{
		{
			metric: &Metric{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost"},
				Fields: map[string]float64{"count": 1}},
			out: `req_count{host="localhost"} 1 1667123357000`,
		},
		{
			metric: &Metric{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost", "region": "us-west-2"},
				Fields: map[string]float64{"count": 1, "max": 10}},
			out: `req_count{host="localhost",region="us-west-2"} 1 1667123357000
req_max{host="localhost",region="us-west-2"} 10 1667123357000`,
		},
	}
	assert := assert.New(t)
	for _, tc := range cases {
		line := tc.metric.EncodePromLines()
		assert.Equal(line, tc.out)
	}
}
