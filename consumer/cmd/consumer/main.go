package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/consumer/internal/handler"
	"github.com/ryan-dayrit/pdf-extractor-pipeline/consumer/internal/nlp"
)

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	nlpClient := nlp.NewClient()
	h := handler.New(nlpClient)
	cgHandler := handler.NewConsumerGroupHandler(h)

	topic := "pdf-events"
	groupID := "pdf-extractor-group"

	var cg sarama.ConsumerGroup
	var err error
	for i := 0; i < 15; i++ {
		cg, err = sarama.NewConsumerGroup([]string{broker}, groupID, cfg)
		if err == nil {
			break
		}
		log.Printf("Kafka not ready, retrying in 3s... (%v)", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to create consumer group: %v", err)
	}
	defer cg.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	log.Printf("consumer started, listening on topic %s", topic)
	for {
		if err := cg.Consume(ctx, []string{topic}, cgHandler); err != nil {
			log.Printf("consume error: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
		time.Sleep(time.Second)
	}
	log.Println("consumer stopped")
}
