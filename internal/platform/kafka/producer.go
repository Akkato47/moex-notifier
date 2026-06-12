package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
	)
	if err != nil {
		return nil, fmt.Errorf("new kafka client: %w", err)
	}
	return &Producer{client: cl, topic: topic}, nil
}

func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	res := p.client.ProduceSync(ctx, &kgo.Record{Key: key, Value: value})
	return res.FirstErr()
}

func (p *Producer) Close() error {
	p.client.Close()
	return nil
}
