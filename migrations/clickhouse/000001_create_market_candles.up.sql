CREATE TABLE IF NOT EXISTS market_candles
(
    ticker    LowCardinality(String),
    timeframe LowCardinality(String),
    ts        DateTime,
    open      Float64,
    high      Float64,
    low       Float64,
    close     Float64,
    volume    UInt64
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (ticker, timeframe, ts);
