package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/consumer/internal/nlp"
)

type Handler struct {
	nlpClient      *nlp.Client
	grpcServiceURL string
	httpClient     *http.Client
}

func New(nlpClient *nlp.Client) *Handler {
	url := os.Getenv("GRPC_SERVICE_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	return &Handler{
		nlpClient:      nlpClient,
		grpcServiceURL: url,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

type ConsumerGroupHandler struct {
	handler *Handler
}

func NewConsumerGroupHandler(h *Handler) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{handler: h}
}

func (cgh *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (cgh *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (cgh *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		cgh.handler.processMessage(msg.Value)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *Handler) processMessage(data []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("failed to unmarshal event: %v", err)
		return
	}

	eventType, _ := event["event"].(string)
	log.Printf("received event: %s", eventType)

	switch eventType {
	case "document.uploaded":
		docID, _ := event["document_id"].(string)
		log.Printf("document uploaded: %s, waiting for data points", docID)

	case "datapoints.submitted":
		h.handleDatapointsSubmitted(event)
	}
}

func (h *Handler) handleDatapointsSubmitted(event map[string]interface{}) {
	docID, _ := event["document_id"].(string)
	docContent, _ := event["document_content"].(string)

	rawDPs, _ := event["data_points"].([]interface{})
	dataPoints := make([]nlp.DataPoint, 0, len(rawDPs))
	for _, dp := range rawDPs {
		dpMap, ok := dp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := dpMap["name"].(string)
		desc, _ := dpMap["description"].(string)
		dataPoints = append(dataPoints, nlp.DataPoint{Name: name, Description: desc})
	}

	log.Printf("extracting data points for document %s", docID)
	results, err := h.nlpClient.ExtractDataPoints(docContent, dataPoints)
	if err != nil {
		log.Printf("NLP extraction failed for %s: %v", docID, err)
		h.updateResults(docID, nil, "failed")
		return
	}

	h.updateResults(docID, results, "completed")
}

func (h *Handler) updateResults(docID string, results []nlp.ExtractionResult, status string) {
	payload := map[string]interface{}{
		"results": results,
		"status":  status,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}

	url := fmt.Sprintf("%s/api/documents/%s/results", h.grpcServiceURL, docID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		log.Printf("create request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("update results error for %s: %v", docID, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("updated results for %s: status %d", docID, resp.StatusCode)
}
