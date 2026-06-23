package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Rule struct {
	ID             uuid.UUID
	Ticker         string
	ConditionType  string
	ConditionValue float64
	ChatID         int64
}

type RuleRepo struct {
	pool *pgxpool.Pool
}

func NewRuleRepo(pool *pgxpool.Pool) *RuleRepo {
	return &RuleRepo{pool: pool}
}

func (r *RuleRepo) GetActiveRules(ctx context.Context) ([]Rule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, ticker, condition_type, condition_value, chat_id
		 FROM alert_rules
		 WHERE is_active = true`,
	)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.ID, &rule.Ticker, &rule.ConditionType, &rule.ConditionValue, &rule.ChatID); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
