package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Akkato47/moex-notifier/internal/ingester/features/candle/domain"
)

type moexResponse struct {
	Candles struct {
		Columns []string `json:"columns"`
		Data    [][]any  `json:"data"`
	} `json:"candles"`
}

type MoexClient struct {
	client  *http.Client
	baseURL string
}

func NewMoexClient(client *http.Client, baseURL string) *MoexClient {
	return &MoexClient{client: client, baseURL: baseURL}
}

func (c *MoexClient) GetCandlesByPeriod(ctx context.Context, ticker string, from, to time.Time) ([]domain.Candle, error) {
	endpoint := fmt.Sprintf(
		"%s/engines/stock/markets/shares/boards/TQBR/securities/%s/candles.json",
		c.baseURL, ticker,
	)

	q := url.Values{}
	q.Set("interval", "1")
	q.Set("from", from.Format("2006-01-02 15:04:05"))
	q.Set("till", to.Format("2006-01-02 15:04:05"))
	q.Set("iss.meta", "off")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var raw moexResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	idx := make(map[string]int, len(raw.Candles.Columns))
	for i, col := range raw.Candles.Columns {
		idx[col] = i
	}

	moscowLoc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, fmt.Errorf("load moscow tz: %w", err)
	}

	candles := make([]domain.Candle, 0, len(raw.Candles.Data))
	for _, row := range raw.Candles.Data {
		ts, err := time.ParseInLocation("2006-01-02 15:04:05", row[idx["begin"]].(string), moscowLoc)
		if err != nil {
			return nil, fmt.Errorf("parse candle time: %w", err)
		}
		candles = append(candles, domain.Candle{
			Ticker:    ticker,
			Timeframe: "1m",
			Ts:        ts,
			Open:      row[idx["open"]].(float64),
			High:      row[idx["high"]].(float64),
			Low:       row[idx["low"]].(float64),
			Close:     row[idx["close"]].(float64),
			Volume:    uint64(row[idx["volume"]].(float64)),
		})
	}

	return candles, nil
}
