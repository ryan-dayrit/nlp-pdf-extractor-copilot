package kafka

import (
	"encoding/json"
	"log"
	"os"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewNoopProducer() *Producer {
	return &Producer{topic: "pdf-events"}
}

func NewProducer() (*Producer, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll

	p, err := sarama.NewSyncProducer([]string{broker}, cfg)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: p, topic: "pdf-events"}, nil
}

func (p *Producer) Publish(event map[string]interface{}) error {
	if p.producer == nil {
		log.Printf("kafka producer not initialized, skipping event: %v", event["event"])
		return nil
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(data),
	}
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("kafka publish error: %v", err)
	}
	return err
}

func (p *Producer) Close() {
	if p.producer != nil {
		p.producer.Close()
	}
}
