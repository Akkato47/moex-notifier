import { Kafka, Consumer } from 'kafkajs';

export function createConsumer(brokers: string[], groupId: string): Consumer {
  const kafka = new Kafka({ brokers });
  return kafka.consumer({ groupId });
}
