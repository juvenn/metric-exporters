package influx

import (
	"testing"
	"time"

	exporters "github.com/juvenn/metric-exporters"
	"github.com/stretchr/testify/assert"
)

func TestEncodeLine(t *testing.T) {
	cases := []struct {
		metric *exporters.Datum
		out    string
	}{
		{
			metric: &exporters.Datum{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost"},
				Fields: map[string]float64{"count": 1}},
			out: "req,host=localhost count=1 1667123357",
		},
		{
			metric: &exporters.Datum{Name: "req", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost", "region": "us-west-2"},
				Fields: map[string]float64{"count": 1, "max": 10}},
			out: "req,host=localhost,region=us-west-2 count=1,max=10 1667123357",
		},
		{
			// name with labels
			metric: &exporters.Datum{Name: "req,method=POST", Type: "counter", Time: time.Unix(1667123357, 0),
				Labels: map[string]string{"host": "localhost", "region": "us-west-2"},
				Fields: map[string]float64{"count": 1, "max": 10}},
			out: "req,method=POST,host=localhost,region=us-west-2 count=1,max=10 1667123357",
		},
	}
	assert := assert.New(t)
	for _, tc := range cases {
		line := encodeLine(tc.metric, "s")
		assert.Equal(line, tc.out)
	}
}

func TestNewEmitter(t *testing.T) {
	assert := assert.New(t)
	em, err := NewV1Emitter("http://127.0.0.1/write", "req", WithRetentionPolicy("daily"), WithPrecision("ms"))
	if err != nil {
		t.Fatalf("%+v\n", err)
	}
	assert.Equal("http://127.0.0.1/write?db=req&precision=ms&rp=daily", em.writeUrl.String())

	em, err = NewV2Emitter("http://127.0.0.1/api/v2/write", "req", WithOrg("ACME"), WithPrecision("ms"))
	if err != nil {
		t.Fatalf("%+v\n", err)
	}
	assert.Equal("http://127.0.0.1/api/v2/write?bucket=req&org=ACME&precision=ms", em.writeUrl.String())
}
