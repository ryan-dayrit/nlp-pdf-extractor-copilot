package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
)

// KafkaMessage represents the message format from the document-uploads topic.
type KafkaMessage struct {
	DocumentID   string   `json:"document_id"`
	Filename     string   `json:"filename"`
	PDFDataB64   string   `json:"pdf_data_base64"`
	DataPoints   []string `json:"data_points"`
}

// DataPointsPayload is the body sent to the gRPC service.
type DataPointsPayload struct {
	Results map[string]string `json:"results"`
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler.
type ConsumerGroupHandler struct{}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session setup")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session cleanup")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in ConsumeClaim: %v", r)
		}
	}()

	for msg := range claim.Messages() {
		log.Printf("Received message: partition=%d offset=%d", msg.Partition, msg.Offset)

		var km KafkaMessage
		if err := json.Unmarshal(msg.Value, &km); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		log.Printf("Processing document_id=%s filename=%s", km.DocumentID, km.Filename)

		results, err := callNLPService(km.PDFDataB64, km.DataPoints)
		if err != nil {
			log.Printf("NLP service error for document_id=%s: %v — marking message consumed", km.DocumentID, err)
			session.MarkMessage(msg, "")
			continue
		}

		log.Printf("NLP extraction complete for document_id=%s, sending results to gRPC service", km.DocumentID)

		if err := sendResultsToGRPCService(km.DocumentID, results); err != nil {
			log.Printf("Failed to send results to gRPC service for document_id=%s: %v", km.DocumentID, err)
		} else {
			log.Printf("Successfully updated document_id=%s", km.DocumentID)
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

func sendResultsToGRPCService(documentID string, results map[string]string) error {
	grpcServiceURL := getEnv("GRPC_SERVICE_URL", "http://grpc-service:8080")
	url := fmt.Sprintf("%s/documents/%s/datapoints", grpcServiceURL, documentID)

	payload := DataPointsPayload{Results: results}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("POST to gRPC service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gRPC service returned status %d", resp.StatusCode)
	}

	log.Printf("gRPC service accepted results for document_id=%s (status %d)", documentID, resp.StatusCode)
	return nil
}

// getEnvHandler allows handler.go to read env vars without importing os directly
// (uses the getEnv helper defined in main.go).
var _ = os.Getenv // ensure os is used
