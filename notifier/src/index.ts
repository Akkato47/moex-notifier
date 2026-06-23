import { getEnv } from './config/env.config';
import { createConsumer } from './kafka/consumer';
import { RedisService } from './services/redis.service';
import { TelegramService } from './services/telegram.service';
import { AlertService } from './services/alert.service';
import { AlertEvent } from './types/alert.types';

async function main() {
  const env = getEnv();

  const redis = new RedisService(env.REDIS_HOST, env.REDIS_PORT);
  const telegram = new TelegramService(env.TELEGRAM_TOKEN);
  const alertService = new AlertService(redis, telegram, env.ALERT_COOLDOWN_SEC);

  const consumer = createConsumer(env.KAFKA_BROKERS.split(','), env.KAFKA_GROUP_ID);
  await consumer.connect();
  await consumer.subscribe({ topic: env.KAFKA_TOPIC, fromBeginning: false });

  console.log('Notifier started, waiting for alerts...');

  await consumer.run({
    eachMessage: async ({ message }) => {
      if (!message.value) return;

      let event: AlertEvent;
      try {
        event = JSON.parse(message.value.toString());
      } catch {
        console.error('Failed to parse message:', message.value.toString());
        return;
      }

      try {
        await alertService.handle(event);
      } catch (err) {
        console.error(`Failed to handle alert rule_id=${event.rule_id}:`, err);
      }
    },
  });

  const shutdown = async () => {
    await consumer.disconnect();
    await redis.quit();
    process.exit(0);
  };

  process.on('SIGTERM', shutdown);
  process.on('SIGINT', shutdown);
}

main().catch((err) => {
  console.error('Fatal:', err);
  process.exit(1);
});
