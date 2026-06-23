import { AlertEvent } from '../types/alert.types';
import { RedisService } from './redis.service';
import { TelegramService } from './telegram.service';

export class AlertService {
  constructor(
    private readonly redis: RedisService,
    private readonly telegram: TelegramService,
    private readonly cooldownSec: number,
  ) {}

  async handle(event: AlertEvent): Promise<void> {
    const isDup = await this.redis.isDuplicate(event.rule_id, this.cooldownSec);
    if (isDup) {
      console.log(`Duplicate skipped: rule_id=${event.rule_id}`);
      return;
    }

    await this.telegram.sendAlert(
      event.chat_id,
      event.ticker,
      event.condition_type,
      event.condition_value,
      event.triggered_value,
    );

    console.log(`Alert sent: rule_id=${event.rule_id} ticker=${event.ticker} chat_id=${event.chat_id}`);
  }
}
