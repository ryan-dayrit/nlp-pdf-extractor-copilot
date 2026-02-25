package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	groupID := getEnv("KAFKA_GROUP_ID", "pdf-extractor-consumer")
	topic := "document-uploads"

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	var consumerGroup sarama.ConsumerGroup
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		consumerGroup, err = sarama.NewConsumerGroup(brokers, groupID, config)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to Kafka (attempt %d/10): %v", attempt, err)
		if attempt < 10 {
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer group after 10 attempts: %v", err)
	}
	defer consumerGroup.Close()

	log.Printf("Connected to Kafka brokers: %v", brokers)

	handler := &ConsumerGroupHandler{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, handler); err != nil {
				log.Printf("Consumer group error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Printf("Consumer started. Listening on topic: %s", topic)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Shutting down consumer...")
}
