package agent

import (
	"fmt"

	"github.com/waynenilsen/waynebot/internal/db"
)

// BudgetChecker checks whether a persona is within its token budget.
type BudgetChecker struct {
	DB *db.DB
}

// NewBudgetChecker creates a BudgetChecker.
func NewBudgetChecker(d *db.DB) *BudgetChecker {
	return &BudgetChecker{DB: d}
}

// WithinBudget returns true if the persona has used fewer tokens than maxTokensPerHour in the last hour.
func (bc *BudgetChecker) WithinBudget(personaID int64, maxTokensPerHour int) (bool, error) {
	if maxTokensPerHour <= 0 {
		return true, nil
	}

	var total int64
	err := bc.DB.SQL.QueryRow(
		`SELECT COALESCE(SUM(prompt_tokens + completion_tokens), 0)
		 FROM llm_calls
		 WHERE persona_id = ? AND created_at >= datetime('now', '-1 hour')`,
		personaID,
	).Scan(&total)
	if err != nil {
		return false, fmt.Errorf("query token usage: %w", err)
	}

	return total < int64(maxTokensPerHour), nil
}
