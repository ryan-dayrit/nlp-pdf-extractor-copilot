package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/ryan-dayrit/nlp-pdf-extractor/grpc-service/kafka"
	"github.com/ryan-dayrit/nlp-pdf-extractor/grpc-service/pb"
	"github.com/ryan-dayrit/nlp-pdf-extractor/grpc-service/server"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	grpcPort := getEnv("GRPC_PORT", "50051")
	httpPort := getEnv("HTTP_PORT", "8080")

	// Kafka producer — optional; service stays up even if Kafka is unavailable.
	var producer *kafka.Producer
	if p, err := kafka.NewProducer(kafkaBrokers); err != nil {
		log.Printf("warning: kafka unavailable (%v) — uploads will not be published", err)
	} else {
		producer = p
		defer producer.Close()
	}

	srv := server.NewServer(producer)

	// gRPC server on grpcPort
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Fatalf("grpc: listen: %v", err)
		}
		gs := grpc.NewServer()
		pb.RegisterExtractorServiceServer(gs, srv)
		log.Printf("gRPC server listening on :%s", grpcPort)
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("grpc: serve: %v", err)
		}
	}()

	// HTTP/REST server on httpPort (blocks main goroutine)
	mux := srv.NewHTTPMux()
	log.Printf("HTTP server listening on :%s", httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), mux); err != nil {
		log.Fatalf("http: serve: %v", err)
	}
}
