package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestAssembleContextPriorityOrdering(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "testbot", "You are helpful.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "general", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello message")
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		History:   history,
	})

	if len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}
	if msgs[0].OfSystem == nil {
		t.Fatal("expected first message to be a system message")
	}

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "You are helpful.") {
		t.Error("system message should contain persona prompt")
	}

	if budget.HistoryMessages != 3 {
		t.Errorf("expected 3 history messages, got %d", budget.HistoryMessages)
	}
}

func TestAssembleContextBudgetExhaustion(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "budgetbot", "Short prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "budget-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	for i := 0; i < 20; i++ {
		_, err := model.CreateMessage(d, ch.ID, 999, "human", "alice", strings.Repeat("word ", 100))
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	_, budget := assembler.AssembleContext(AssembleInput{
		Persona:    persona,
		ChannelID:  ch.ID,
		History:    history,
		TokenLimit: 300,
	})

	if !budget.Exhausted {
		t.Error("expected budget to be exhausted with tight token limit")
	}
	if budget.HistoryMessages >= 20 {
		t.Errorf("expected fewer than 20 history messages, got %d", budget.HistoryMessages)
	}
	if budget.HistoryMessages == 0 {
		t.Error("expected at least some history messages")
	}
}

func TestAssembleContextBasic(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "basicbot", "You are a bot.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "basic", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hi there")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, _ := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		History:   history,
	})

	if len(msgs) != 2 { // system + 1 history
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestAssembleContextWithProjectContext(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "projbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "proj-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	proj, err := model.CreateProject(d, "myproject", t.TempDir(), "A cool project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	if msgs[0].OfSystem == nil {
		t.Fatal("expected system message")
	}

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "myproject") {
		t.Error("system message should contain project name")
	}
	if !strings.Contains(sysContent, "A cool project") {
		t.Error("system message should contain project description")
	}
	if budget.ProjectTokens == 0 {
		t.Error("expected non-zero project tokens")
	}
}

func TestAssembleContextWithAgentsMd(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "agentsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "agents-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	projDir := t.TempDir()
	err = os.WriteFile(filepath.Join(projDir, "AGENTS.md"), []byte("Always use gofmt.\nPrefer table-driven tests."), 0644)
	if err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	proj, err := model.CreateProject(d, "agentsproject", projDir, "A project with AGENTS.md")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "## Project Instructions (AGENTS.md)") {
		t.Error("system message should contain AGENTS.md header")
	}
	if !strings.Contains(sysContent, "Always use gofmt.") {
		t.Error("system message should contain AGENTS.md content")
	}
	if budget.AgentsmdTokens == 0 {
		t.Error("expected non-zero AgentsmdTokens in budget")
	}
}

func TestAssembleContextWithoutAgentsMd(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "noagentsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "no-agents-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	proj, err := model.CreateProject(d, "plainproject", t.TempDir(), "No agents file")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	_, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	if budget.AgentsmdTokens != 0 {
		t.Errorf("expected 0 AgentsmdTokens without AGENTS.md, got %d", budget.AgentsmdTokens)
	}
}

func TestAssembleContextWithProjectDocuments(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "docsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "docs-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	projDir := t.TempDir()
	for _, dir := range []string{"erd", "prd", "decisions"} {
		if err := os.MkdirAll(filepath.Join(projDir, dir), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte("# ERD\nUsers -> Posts -> Comments"), 0644); err != nil {
		t.Fatalf("write erd/main.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projDir, "prd", "main.md"), []byte("# PRD\nBuild a chat app with AI agents."), 0644); err != nil {
		t.Fatalf("write prd/main.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projDir, "decisions", "log.md"), []byte("## Use SQLite\nWe chose SQLite for simplicity.\n---\n## Use Go\nGo is fast and simple."), 0644); err != nil {
		t.Fatalf("write decisions/log.md: %v", err)
	}

	proj, err := model.CreateProject(d, "docsproj", projDir, "A project with docs")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	sysContent := msgs[0].OfSystem.Content.OfString.Value

	if !strings.Contains(sysContent, "## Project Documents") {
		t.Error("system message should contain Project Documents header")
	}
	if !strings.Contains(sysContent, "Users -> Posts -> Comments") {
		t.Error("system message should contain ERD content")
	}
	if !strings.Contains(sysContent, "Build a chat app with AI agents") {
		t.Error("system message should contain PRD content")
	}
	if !strings.Contains(sysContent, "Use SQLite") {
		t.Error("system message should contain decisions content")
	}
	if budget.DocumentTokens == 0 {
		t.Error("expected non-zero DocumentTokens in budget")
	}
}

func TestAssembleContextWithoutProjectDocuments(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "nodocsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "nodocs-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	proj, err := model.CreateProject(d, "nodocsproject", t.TempDir(), "No docs")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	_, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	if budget.DocumentTokens != 0 {
		t.Errorf("expected 0 DocumentTokens without doc dirs, got %d", budget.DocumentTokens)
	}
}

func TestTruncateDecisions(t *testing.T) {
	var sb strings.Builder
	for i := 1; i <= 25; i++ {
		if i > 1 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("## Decision %d\nSome details about decision %d.\n", i, i))
	}

	result := truncateDecisions(sb.String(), 20)

	if strings.Contains(result, "Decision 5\n") {
		t.Error("truncated decisions should not contain early entries")
	}
	if !strings.Contains(result, "Decision 6") {
		t.Error("truncated decisions should contain entry 6")
	}
	if !strings.Contains(result, "Decision 25") {
		t.Error("truncated decisions should contain the last entry")
	}
}

func TestAssembleContextAgentsmdTruncation(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "truncbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "trunc-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	projDir := t.TempDir()
	bigContent := strings.Repeat("x", 20_000)
	if err := os.WriteFile(filepath.Join(projDir, "AGENTS.md"), []byte(bigContent), 0644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	proj, err := model.CreateProject(d, "truncproj", projDir, "truncation test")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	sysContent := msgs[0].OfSystem.Content.OfString.Value

	if !strings.Contains(sysContent, "## Project Instructions (AGENTS.md)") {
		t.Error("system message should contain AGENTS.md header even when truncated")
	}
	if budget.AgentsmdTokens > maxAgentsmdChars/4+100 {
		t.Errorf("AgentsmdTokens = %d, expected roughly %d or less", budget.AgentsmdTokens, maxAgentsmdChars/4)
	}
	if budget.AgentsmdTokens == 0 {
		t.Error("expected non-zero AgentsmdTokens for large AGENTS.md")
	}
}

func TestAssembleContextDocsBudget(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "budgetdocsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "docsbudget-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "erd"), 0755)
	os.MkdirAll(filepath.Join(projDir, "prd"), 0755)

	bigERD := strings.Repeat("E", 20_000)
	bigPRD := strings.Repeat("P", 20_000)
	os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte(bigERD), 0644)
	os.WriteFile(filepath.Join(projDir, "prd", "main.md"), []byte(bigPRD), 0644)

	proj, err := model.CreateProject(d, "budgetdocsproj", projDir, "docs budget test")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{}

	msgs, budget := assembler.AssembleContext(AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	sysContent := msgs[0].OfSystem.Content.OfString.Value

	if !strings.Contains(sysContent, "## Project Documents") {
		t.Error("system message should contain Project Documents header")
	}

	maxExpectedTokens := maxDocumentChars/4 + 100
	if budget.DocumentTokens > maxExpectedTokens {
		t.Errorf("DocumentTokens = %d, expected at most ~%d", budget.DocumentTokens, maxExpectedTokens)
	}
	if budget.DocumentTokens == 0 {
		t.Error("expected non-zero DocumentTokens")
	}
}

func TestEstimateTokens(t *testing.T) {
	got := EstimateTokens("Hello world!") // 12 chars = 3 tokens
	if got != 3 {
		t.Errorf("expected 3 tokens, got %d", got)
	}
}
