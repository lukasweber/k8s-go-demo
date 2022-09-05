package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"sync"

	pb "calculator-service/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

var (
	operationsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calc_operations_total",
		Help: "The total number of operations",
	})
)

// server is used to implement proto.CalculatorServer.
type server struct {
	pb.UnimplementedCalculatorServer
}

// CalculatePrimeNumbers implements proto.CalculatorServer
func (s *server) CalculatePrimeNumbers(ctx context.Context, in *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	log.Printf("Received: %d, %d", in.GetFrom(), in.GetTo())
	numbers := getPrimeNumbers(in.GetFrom(), in.GetTo())
	operationsCounter.Inc()
	hostname, _ := os.Hostname()
	return &pb.CalculateResponse{Count: int32(len(numbers)), Hostname: hostname}, nil
}

func main() {
	flag.Parse()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// GRPC Service
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterCalculatorServer(s, &server{})
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
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

func getPrimeNumbers(num1, num2 int32) []int32 {
	out := []int32{}
	if num1 < 2 || num2 < 2 {
		fmt.Println("Numbers must be greater than 2.")
		return out
	}
	for num1 <= num2 {
		isPrime := true
		for i := 2; i <= int(math.Sqrt(float64(num1))); i++ {
			if num1%int32(i) == 0 {
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
