package influx

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	exporters "github.com/juvenn/metric-exporters"
	"github.com/rcrowley/go-metrics"
)

var (
	influxAddr = "127.0.0.1:8080"
	influxUrl  = fmt.Sprintf("http://%s/write", influxAddr)
)

func Example() {
	em, err := NewV1Emitter(influxUrl, "req")
	if err != nil {
		fmt.Printf("Error %+v\n", err)
	}
	srv := startInfluxServer(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("path:", req.URL.String())
		bstr, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("Error %+v\n", err)
		}
		num := strings.Count(string(bstr), "\n")
		fmt.Println("count:", num)
		w.WriteHeader(204)
	})
	defer srv.Close()
	reg := metrics.NewRegistry()
	rep, err := exporters.NewReporter(reg, 1*time.Second, exporters.WithEmitters(em))
	if err != nil {
		fmt.Printf("Error %+v\n", err)
	}

	rep.Start()
	counter := metrics.NewCounter()
	reg.Register("req", counter)
	counter.Inc(1)
	time.Sleep(1800 * time.Millisecond)
	rep.Close()
	// Output:
	// path: /write?db=req&precision=s
	// count: 1
	// path: /write?db=req&precision=s
	// count: 1
}

func startInfluxServer(handle func(w http.ResponseWriter, req *http.Request)) *httptest.Server {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(handle))
	li, err := net.Listen("tcp", influxAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on tcp %s: %v\n", influxAddr, err))
	}
	srv.Listener = li
	srv.Start()
	return srv
}
