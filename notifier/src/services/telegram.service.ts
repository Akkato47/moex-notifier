import TelegramBot from 'node-telegram-bot-api';

export class TelegramService {
  private bot: TelegramBot;

  constructor(token: string) {
    this.bot = new TelegramBot(token);
  }

  async sendAlert(chatId: number, ticker: string, conditionType: string, conditionValue: number, triggeredValue: number): Promise<void> {
    const direction = conditionType.includes('below') ? 'ниже' : 'выше';
    const subject = conditionType.startsWith('ma') ? 'MA(20)' : 'Цена';
    const text = `🔔 *${ticker}*\n${subject} ${direction} ${conditionValue}\nТекущее значение: *${triggeredValue.toFixed(2)}*`;

    await this.bot.sendMessage(chatId, text, { parse_mode: 'Markdown' });
  }
}
