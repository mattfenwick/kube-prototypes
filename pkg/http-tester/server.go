package http_tester

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
)

func RunServer(args []string) {
	port, err := strconv.ParseInt(args[0], 10, 32)
	doOrDie(err)

	log.SetLevel(log.DebugLevel)

	// Prometheus and http setup
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", port)
	go func() {
		log.Infof("starting HTTP server on port %d", port)
		http.ListenAndServe(addr, nil)
	}()

	respNumber := 0
	SetupHTTPServer(func(request *Request) *Response {
		resp := &Response{
			Request:         request,
			ResponseNumber:  respNumber,
			ResponseMessage: fmt.Sprintf("this is the %d response", respNumber),
		}
		log.Infof("handling request with response: %s", resp.JSONString(false))
		respNumber++
		return resp
	})

	stop := make(chan struct{})
	<-stop
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	log.Errorf("HTTPResponder not found from request %+v", r)
	http.NotFound(w, r)
}

func Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	http.Error(w, err.Error(), statusCode)
}

func SetupHTTPServer(handler func(request *Request) *Response) {
	http.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("received request: %s", r.URL.Path)
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("unable to read body for POST: %s", err.Error())
				Error(w, r, err, 400)
				return
			}
			var req Request
			err = json.Unmarshal(body, &req)
			if err != nil {
				log.Errorf("unable to ummarshal JSON for POST: %s", err.Error())
				Error(w, r, err, 400)
				return
			}
			resp := handler(&req)
			bytes, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				log.Errorf("unable to marshal json for POST: %s", err)
				Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			_, err = fmt.Fprint(w, string(bytes))
			doOrDie(err)
		default:
			NotFound(w, r)
		}
	})
}
