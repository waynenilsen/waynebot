package agent

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func testPersona(id int64, name string, cooldown int) model.Persona {
	return model.Persona{
		ID:           id,
		Name:         name,
		CooldownSecs: cooldown,
	}
}

func msg(authorID int64, authorType, content string) model.Message {
	return model.Message{
		AuthorID:   authorID,
		AuthorType: authorType,
		Content:    content,
	}
}

func TestShouldRespondNoMessages(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 30)

	if dm.ShouldRespond(p, 1, nil) {
		t.Error("expected false for empty messages")
	}
}

func TestShouldRespondAllSelfMessages(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 0)

	msgs := []model.Message{
		msg(1, "agent", "hello"),
		msg(1, "agent", "world"),
	}
	if dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected false when all messages from self")
	}
}

func TestShouldRespondHumanMessage(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 0)

	msgs := []model.Message{
		msg(99, "human", "hey bot"),
	}
	if !dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected true for human message")
	}
}

func TestShouldRespondMixedMessages(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 0)

	msgs := []model.Message{
		msg(1, "agent", "I said something"),
		msg(99, "human", "user reply"),
	}
	if !dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected true when messages include non-self")
	}
}

func TestShouldRespondCooldownBlocks(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 3600) // 1 hour cooldown

	dm.RecordResponse(1, 1)

	msgs := []model.Message{
		msg(99, "human", "hello"),
	}
	if dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected false during cooldown")
	}
}

func TestShouldRespondMentionOverridesCooldown(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 3600)

	dm.RecordResponse(1, 1)

	msgs := []model.Message{
		msg(99, "human", "hey @bot can you help?"),
	}
	if !dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected true when @mentioned even during cooldown")
	}
}

func TestShouldRespondMentionOverridesSelfMessages(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 0)

	msgs := []model.Message{
		msg(1, "agent", "@bot reply to yourself"),
	}
	if !dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected true when @mentioned even if all self")
	}
}

func TestShouldRespondMentionCaseInsensitive(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "MyBot", 0)

	msgs := []model.Message{
		msg(99, "human", "hey @mybot please help"),
	}
	if !dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected true for case-insensitive mention")
	}
}

func TestShouldRespondDifferentChannelNoCooldown(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 3600)

	dm.RecordResponse(1, 1)

	msgs := []model.Message{
		msg(99, "human", "hello"),
	}
	// channel 2 should not be affected by channel 1's cooldown
	if !dm.ShouldRespond(p, 2, msgs) {
		t.Error("expected true for different channel")
	}
}

func TestRecordResponseStartsCooldown(t *testing.T) {
	dm := NewDecisionMaker()
	p := testPersona(1, "bot", 3600)

	msgs := []model.Message{
		msg(99, "human", "hello"),
	}

	if !dm.ShouldRespond(p, 1, msgs) {
		t.Fatal("expected true before cooldown")
	}

	dm.RecordResponse(1, 1)

	if dm.ShouldRespond(p, 1, msgs) {
		t.Error("expected false after RecordResponse")
	}
}
