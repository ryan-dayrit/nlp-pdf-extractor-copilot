// Package kafka provides a Sarama-based synchronous Kafka producer.
package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

const topicDocumentUploads = "document-uploads"

// Producer wraps a Sarama SyncProducer.
type Producer struct {
	sp sarama.SyncProducer
}

// documentUploadEvent is the JSON payload published to document-uploads.
type documentUploadEvent struct {
	DocumentID    string   `json:"document_id"`
	Filename      string   `json:"filename"`
	PDFDataBase64 string   `json:"pdf_data_base64"`
	DataPoints    []string `json:"data_points"`
}

// NewProducer creates a synchronous Kafka producer connected to brokers
// (comma-separated list, e.g. "kafka:9092").
func NewProducer(brokers string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	sp, err := sarama.NewSyncProducer(strings.Split(brokers, ","), cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: new sync producer: %w", err)
	}
	return &Producer{sp: sp}, nil
}

// PublishDocumentUpload sends a document-upload event to the document-uploads topic.
func (p *Producer) PublishDocumentUpload(docID, filename, pdfBase64 string, dataPoints []string) error {
	evt := documentUploadEvent{
		DocumentID:    docID,
		Filename:      filename,
		PDFDataBase64: pdfBase64,
		DataPoints:    dataPoints,
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("kafka: marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topicDocumentUploads,
		Value: sarama.ByteEncoder(payload),
	}
	_, _, err = p.sp.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("kafka: send message: %w", err)
	}
	log.Printf("kafka: published upload event doc_id=%s", docID)
	return nil
}

// Close shuts down the producer gracefully.
func (p *Producer) Close() error {
	return p.sp.Close()
}
