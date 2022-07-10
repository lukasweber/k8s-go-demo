package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CalcRequest struct {
	From int
	To   int
}

type CalcResponse struct {
	Result   []int
	Hostname string
}

func getPrimeNumbers(num1, num2 int) []int {
	out := []int{}
	if num1 < 2 || num2 < 2 {
		fmt.Println("Numbers must be greater than 2.")
		return out
	}
	for num1 <= num2 {
		isPrime := true
		for i := 2; i <= int(math.Sqrt(float64(num1))); i++ {
			if num1%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			out = append(out, num1)
		}
		num1++
	}
	return out
}

var (
	operationsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calc_operations_total",
		Help: "The total number of operations",
	})
)

func handleCalcRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Request: calcPrime")
	// parse request
	var requestBody CalcRequest
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	if err := json.Unmarshal(body, &requestBody); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	// create response
	nums := getPrimeNumbers(requestBody.From, requestBody.To)
	hostname, _ := os.Hostname()
	json.NewEncoder(w).Encode(CalcResponse{Result: nums, Hostname: hostname})
	operationsCounter.Inc()
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// API
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/calc", handleCalcRequest)
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
