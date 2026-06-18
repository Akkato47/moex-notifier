CREATE TABLE IF NOT EXISTS alert_rules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker           VARCHAR(20)  NOT NULL,
    condition_type   VARCHAR(50)  NOT NULL,
    condition_value  FLOAT8       NOT NULL,
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_active_ticker ON alert_rules (is_active, ticker);
