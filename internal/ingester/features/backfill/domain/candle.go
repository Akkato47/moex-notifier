package domain

import "time"

type Candle struct {
	Ticker    string
	Timeframe string
	Ts        time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    uint64
}
