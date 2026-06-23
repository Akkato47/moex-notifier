export interface AlertEvent {
  rule_id: string;
  ticker: string;
  chat_id: number;
  condition_type: string;
  condition_value: number;
  triggered_value: number;
  triggered_at: string;
}
