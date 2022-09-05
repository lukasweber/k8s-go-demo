package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/lukasweber/k8s-go-demo/api-service/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CalcApiResponse struct {
	ApiHostName string
	RpcHostName string
	Count       int32
}

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")

	requestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goapi_requests_total",
		Help: "The total number of requests",
	})

	rpcClient pb.CalculatorClient
)

func handleCalcRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Request: calc")

	from, err := strconv.Atoi(r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "from missing or in bad format", 400)
		return
	}

	to, err := strconv.Atoi(r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "to missing or in bad format", 400)
		return
	}

	hostname, _ := os.Hostname()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req, err := rpcClient.CalculatePrimeNumbers(ctx, &pb.CalculateRequest{
		From: int32(from),
		To:   int32(to),
	})
	if err != nil {
		log.Fatalf("could not call grpc service: %v", err)
	}
	log.Printf("RPC Response received")

	json.NewEncoder(w).Encode(CalcApiResponse{ApiHostName: hostname, RpcHostName: req.GetHostname(), Count: req.GetCount()})
	requestsCounter.Inc()
}

func main() {
	flag.Parse()

	// setup grpc client
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	rpcClient = pb.NewCalculatorClient(conn)

	// setup threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// API
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/prime", handleCalcRequest)
		server := http.Server{
			Addr:    ":8080",
			Handler: mux,
		}
		log.Printf("Api Server listening at %v", server.Addr)
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
