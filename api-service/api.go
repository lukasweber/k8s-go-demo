package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goapi_requests_total",
		Help: "The total number of requests",
	})
)

func handleHostRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Request: hostJson")
	hostname, _ := os.Hostname()
	json.NewEncoder(w).Encode(map[string]string{"hostname": hostname})
	requestsCounter.Inc()
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// API
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/host", handleHostRequest)
		server := http.Server{
			Addr:    ":8080",
			Handler: mux,
		}
		server.ListenAndServe()
		wg.Done()
	}()

	// Prometheus Exporter
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			promhttp.Handler().ServeHTTP(w, r)
		})
		server := http.Server{
			Addr:    ":2121",
			Handler: mux,
		}
		server.ListenAndServe()
		wg.Done()
	}()

	wg.Wait()
}
