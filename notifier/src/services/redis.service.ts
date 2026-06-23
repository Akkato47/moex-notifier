import { Redis } from 'ioredis';

export class RedisService {
  private client: Redis;

  constructor(host: string, port: number) {
    this.client = new Redis({
      host,
      port,
      retryStrategy: (times) => Math.min(times * 50, 2000),
    });

    this.client.on('error', (err) => {
      console.error('Redis error:', err);
    });
  }

  async isDuplicate(ruleId: string, ttlSec: number): Promise<boolean> {
    const key = `alert:${ruleId}`;
    const result = await this.client.set(key, '1', 'EX', ttlSec, 'NX');
    return result === null;
  }

  async quit(): Promise<void> {
    await this.client.quit();
  }
}
