package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/Akkato47/moex-notifier/internal/platform/kafka"
	"github.com/Akkato47/moex-notifier/internal/processor/features/candle/repository"
	indicatorDomain "github.com/Akkato47/moex-notifier/internal/processor/features/indicator/domain"
	ruleRepo "github.com/Akkato47/moex-notifier/internal/processor/features/rule/repository"
)

const maPeriod = 20

type AlertEvent struct {
	RuleID         string    `json:"rule_id"`
	Ticker         string    `json:"ticker"`
	ChatID         int64     `json:"chat_id"`
	ConditionType  string    `json:"condition_type"`
	ConditionValue float64   `json:"condition_value"`
	TriggeredValue float64   `json:"triggered_value"`
	TriggeredAt    time.Time `json:"triggered_at"`
}

type Matcher struct {
	candleReader *repository.CandleReader
	ruleRepo     *ruleRepo.RuleRepo
	producer     *kafka.Producer
	log          *zerolog.Logger
	pollInterval time.Duration
	watermark    map[string]time.Time
	maStates     map[string]*indicatorDomain.MAState
}

func NewMatcher(
	candleReader *repository.CandleReader,
	ruleRepo *ruleRepo.RuleRepo,
	producer *kafka.Producer,
	log *zerolog.Logger,
	pollInterval time.Duration,
) *Matcher {
	return &Matcher{
		candleReader: candleReader,
		ruleRepo:     ruleRepo,
		producer:     producer,
		log:          log,
		pollInterval: pollInterval,
		watermark:    make(map[string]time.Time),
		maStates:     make(map[string]*indicatorDomain.MAState),
	}
}

func (m *Matcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := m.process(ctx); err != nil {
				m.log.Error().Err(err).Msg("process tick failed")
			}
		}
	}
}

func (m *Matcher) process(ctx context.Context) error {
	rules, err := m.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return fmt.Errorf("get rules: %w", err)
	}

	byTicker := make(map[string][]ruleRepo.Rule)
	for _, r := range rules {
		byTicker[r.Ticker] = append(byTicker[r.Ticker], r)
	}

	for ticker, tickerRules := range byTicker {
		wm, ok := m.watermark[ticker]
		if !ok {
			wm = time.Now().Add(-1 * time.Hour)
			m.watermark[ticker] = wm
		}

		candles, err := m.candleReader.GetCandlesAfter(ctx, ticker, wm)
		if err != nil {
			m.log.Error().Err(err).Str("ticker", ticker).Msg("get candles")
			continue
		}
		if len(candles) == 0 {
			continue
		}

		if _, ok := m.maStates[ticker]; !ok {
			m.maStates[ticker] = indicatorDomain.NewMAState(maPeriod)
		}
		ma := m.maStates[ticker]

		lastCandle := candles[len(candles)-1]
		for _, c := range candles {
			ma.Update(c.Close)
		}
		m.watermark[ticker] = lastCandle.Ts

		if !ma.IsReady() {
			m.log.Debug().Str("ticker", ticker).Msg("ma not ready yet")
			continue
		}

		price := lastCandle.Close
		maVal := ma.Value()

		for _, rule := range tickerRules {
			triggered := false
			var triggeredVal float64

			switch rule.ConditionType {
			case "price_below":
				triggered = price < rule.ConditionValue
				triggeredVal = price
			case "price_above":
				triggered = price > rule.ConditionValue
				triggeredVal = price
			case "ma_below":
				triggered = maVal < rule.ConditionValue
				triggeredVal = maVal
			case "ma_above":
				triggered = maVal > rule.ConditionValue
				triggeredVal = maVal
			}

			if !triggered {
				continue
			}

			event := AlertEvent{
				RuleID:         rule.ID.String(),
				Ticker:         ticker,
				ChatID:         rule.ChatID,
				ConditionType:  rule.ConditionType,
				ConditionValue: rule.ConditionValue,
				TriggeredValue: triggeredVal,
				TriggeredAt:    lastCandle.Ts,
			}
			payload, _ := json.Marshal(event)

			if err := m.producer.Publish(ctx, []byte(rule.ID.String()), payload); err != nil {
				m.log.Error().Err(err).Str("rule_id", rule.ID.String()).Msg("publish alert")
				continue
			}
			m.log.Info().
				Str("rule_id", rule.ID.String()).
				Str("ticker", ticker).
				Str("condition", rule.ConditionType).
				Float64("triggered_value", triggeredVal).
				Msg("alert triggered")
		}
	}
	return nil
}
