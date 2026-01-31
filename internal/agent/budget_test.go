package agent

import (
	"testing"
)

func TestWithinBudgetNoCallsReturnsTrue(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	ok, err := bc.WithinBudget(1, 1000)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if !ok {
		t.Error("expected within budget with no calls")
	}
}

func TestWithinBudgetZeroLimitReturnsTrue(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	ok, err := bc.WithinBudget(1, 0)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if !ok {
		t.Error("expected within budget with zero limit (unlimited)")
	}
}

func TestWithinBudgetUnderLimit(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	// Insert an llm_call with 500 total tokens
	_, err := d.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (1, 1, 'test-model', '[]', '{}', 300, 200)`,
	)
	if err != nil {
		t.Fatalf("insert llm_call: %v", err)
	}

	ok, err := bc.WithinBudget(1, 1000)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if !ok {
		t.Error("expected within budget (500 < 1000)")
	}
}

func TestWithinBudgetOverLimit(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	// Insert calls totaling 1500 tokens
	for _, tokens := range [][2]int{{400, 350}, {400, 350}} {
		_, err := d.WriteExec(
			`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
			 VALUES (1, 1, 'test-model', '[]', '{}', ?, ?)`,
			tokens[0], tokens[1],
		)
		if err != nil {
			t.Fatalf("insert llm_call: %v", err)
		}
	}

	ok, err := bc.WithinBudget(1, 1000)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if ok {
		t.Error("expected over budget (1500 >= 1000)")
	}
}

func TestWithinBudgetIsolatesPersonas(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	// Insert 900 tokens for persona 1
	_, err := d.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (1, 1, 'test-model', '[]', '{}', 500, 400)`,
	)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Persona 2 should still be within budget
	ok, err := bc.WithinBudget(2, 1000)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if !ok {
		t.Error("expected persona 2 within budget")
	}
}

func TestWithinBudgetExcludesOldCalls(t *testing.T) {
	d := openTestDB(t)
	bc := NewBudgetChecker(d)

	// Insert a call with a timestamp > 1 hour ago
	_, err := d.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens, created_at)
		 VALUES (1, 1, 'test-model', '[]', '{}', 5000, 5000, datetime('now', '-2 hours'))`,
	)
	if err != nil {
		t.Fatalf("insert old call: %v", err)
	}

	ok, err := bc.WithinBudget(1, 1000)
	if err != nil {
		t.Fatalf("WithinBudget: %v", err)
	}
	if !ok {
		t.Error("expected within budget, old call should be excluded")
	}
}
