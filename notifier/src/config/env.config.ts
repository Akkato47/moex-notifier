import * as dotenv from 'dotenv';

dotenv.config();

export interface EnvConfig {
  TELEGRAM_TOKEN: string;
  KAFKA_BROKERS: string;
  KAFKA_TOPIC: string;
  KAFKA_GROUP_ID: string;
  REDIS_HOST: string;
  REDIS_PORT: number;
  ALERT_COOLDOWN_SEC: number;
}

const config: EnvConfig = {
  TELEGRAM_TOKEN: process.env.TELEGRAM_TOKEN || '',
  KAFKA_BROKERS: process.env.KAFKA_BROKERS || 'localhost:9092',
  KAFKA_TOPIC: process.env.KAFKA_TOPIC || 'alerts.triggered',
  KAFKA_GROUP_ID: process.env.KAFKA_GROUP_ID || 'notifier',
  REDIS_HOST: process.env.REDIS_HOST || 'localhost',
  REDIS_PORT: parseInt(process.env.REDIS_PORT || '6379', 10),
  ALERT_COOLDOWN_SEC: parseInt(process.env.ALERT_COOLDOWN_SEC || '300', 10),
};

export function getEnv(): EnvConfig {
  if (!config.TELEGRAM_TOKEN) {
    throw new Error('TELEGRAM_TOKEN is required');
  }
  return config;
}
