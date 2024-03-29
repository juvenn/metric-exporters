package influx

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	exporters "github.com/juvenn/metric-exporters"
)

func NewV2Emitter(writeUrl string, bucket string, opts ...Option) (*influxEmitter, error) {
	em, err := newEmitter(writeUrl, opts...)
	if err != nil {
		return nil, err
	}
	em.v2 = true
	em.bucket = bucket
	em.params.Set("bucket", bucket)
	if em.org != "" {
		em.params.Set("org", em.org)
	}
	return em, nil
}

func NewV1Emitter(writeUrl, database string, opts ...Option) (*influxEmitter, error) {
	em, err := newEmitter(writeUrl, opts...)
	if err != nil {
		return nil, err
	}
	em.database = database
	em.params.Set("db", database)
	return em, nil
}

func newEmitter(writeUrl string, opts ...Option) (*influxEmitter, error) {
	url, err := url.Parse(writeUrl)
	if err != nil {
		return nil, err
	}
	em := &influxEmitter{
		writeUrl:  url,
		precision: "s",
		params:    url.Query(),
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(em)
	}
	if !validPrecisions[em.precision] {
		return nil, fmt.Errorf("Influx precision must be one of [ns,u,us,ms,s]")
	}
	if em.precision != "" {
		em.params.Set("precision", em.precision)
	}
	return em, nil
}

// Http based influx emitter that supports both v1 and v2 ednpoints.
// See https://docs.influxdata.com/influxdb/v1.8/tools/api/#influxdb-20-api-compatibility-endpoints
type influxEmitter struct {
	writeUrl  *url.URL   // Influx url
	params    url.Values // Influx url params
	username  string
	password  string
	precision string

	database string // v1

	v2        bool   // v2 or not
	authtoken string // v2
	org       string // v2
	bucket    string // v2

	http *http.Client
}

func (this *influxEmitter) Name() string {
	if this.v2 {
		return fmt.Sprintf("influxdb: %s bucket=%s", this.writeUrl.String(), this.bucket)
	} else {
		return fmt.Sprintf("influxdb: %s db=%s", this.writeUrl.String(), this.database)
	}
}

func (this *influxEmitter) Close() error {
	return nil
}

func (this *influxEmitter) Emit(metrics ...*exporters.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	var lines strings.Builder
	for _, metric := range metrics {
		line := metric.EncodeInfluxLine(this.precision)
		lines.WriteString(line)
		lines.WriteString("\n")
	}
	return this.request(bytes.NewBufferString(lines.String()))
}

func (this *influxEmitter) buildUrl() string {
	url := *this.writeUrl
	url.RawQuery = this.params.Encode()
	return url.String()
}

func (this *influxEmitter) request(body io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, this.buildUrl(), body)
	if err != nil {
		return err
	}
	if this.v2 {
		if this.authtoken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", this.authtoken))
		} else if this.username != "" && this.password != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s:%s", this.username, this.password))
		}
	} else {
		if this.username != "" && this.password != "" {
			req.SetBasicAuth(this.username, this.password)
		}
	}
	req.Header.Set("user-agent", "metrics-exporter/0.1.0")
	resp, err := this.http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bstr, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s %s %s %s", http.MethodPost, this.writeUrl, resp.Status, string(bstr))
	}
	return nil
}

type Option func(*influxEmitter)

// ### Common options

// Timestamp precision used to encode metric, can be one of [ns,u,us,ms,s], default to s.
func WithPrecision(p string) Option {
	return func(em *influxEmitter) {
		em.precision = p
	}
}

var (
	validPrecisions = map[string]bool{
		"ns": true,
		"u":  true, // same as us
		"us": true,
		"ms": true,
		"s":  true,
	}
)

// User pass authentication
func WithUserAuth(user, pass string) Option {
	return func(em *influxEmitter) {
		em.username = user
		em.password = pass
	}
}

// Http request timeout, default to 5s.
func WithRequestTimeout(du time.Duration) Option {
	return func(em *influxEmitter) {
		em.http.Timeout = du
	}
}

// ### V2 options

// Influx API token, v2 only.
// See https://docs.influxdata.com/influxdb/v2.4/security/tokens/
func WithAuthToken(token string) Option {
	return func(em *influxEmitter) {
		em.authtoken = token
	}
}

// Org name, v2 only.
func WithOrg(org string) Option {
	return func(em *influxEmitter) {
		em.org = org
	}
}

// ### V1 options

// Retention policy, v1 only
func WithRetentionPolicy(rp string) Option {
	return func(em *influxEmitter) {
		em.params.Set("rp", rp)
	}
}
