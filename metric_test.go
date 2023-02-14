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

func TestReshape(t *testing.T) {
	metric := &Metric{Name: "req.method.GET", Type: "counter", Time: time.Unix(1667123357, 0),
		Labels: map[string]string{"host": "localhost"},
		Fields: map[string]float64{"count": 1}}
	var reshape Reshape = func(m *Metric) *Metric {
		m.Labels["method"] = "GET"
		m.Name = "req"
		return m
	}
	out := reshape(metric)
	assert := assert.New(t)
	assert.Equal("req.method.GET", metric.Name)
	assert.Equal("", metric.Labels["method"])
	assert.Equal("req", out.Name)
	assert.Equal("GET", out.Labels["method"])
}
