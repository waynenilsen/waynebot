package agent

import (
	"strings"
	"sync"
	"time"

	"github.com/waynenilsen/waynebot/internal/model"
)

// DecisionMaker determines whether a persona should respond in a channel.
type DecisionMaker struct {
	mu        sync.Mutex
	cooldowns map[cooldownKey]time.Time
}

type cooldownKey struct {
	PersonaID int64
	ChannelID int64
}

// NewDecisionMaker creates a DecisionMaker.
func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{cooldowns: make(map[cooldownKey]time.Time)}
}

// ShouldRespond returns true if the persona should respond to the given messages in the channel.
// It checks: (1) not all messages are from self, (2) cooldown has elapsed, (3) @mention override.
func (dm *DecisionMaker) ShouldRespond(persona model.Persona, channelID int64, messages []model.Message) bool {
	if len(messages) == 0 {
		return false
	}

	mentioned := isMentioned(persona.Name, messages)

	if allFromSelf(persona.ID, messages) && !mentioned {
		return false
	}

	if mentioned {
		dm.resetCooldown(persona.ID, channelID)
		return true
	}

	if !dm.cooldownElapsed(persona.ID, channelID, persona.CooldownSecs) {
		return false
	}

	return true
}

// RecordResponse records that a persona responded in a channel, starting the cooldown timer.
func (dm *DecisionMaker) RecordResponse(personaID, channelID int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.cooldowns[cooldownKey{personaID, channelID}] = time.Now()
}

func allFromSelf(personaID int64, messages []model.Message) bool {
	for _, m := range messages {
		if m.AuthorType != "agent" || m.AuthorID != personaID {
			return false
		}
	}
	return true
}

func isMentioned(name string, messages []model.Message) bool {
	mention := "@" + strings.ToLower(name)
	for _, m := range messages {
		if strings.Contains(strings.ToLower(m.Content), mention) {
			return true
		}
	}
	return false
}

func (dm *DecisionMaker) cooldownElapsed(personaID, channelID int64, cooldownSecs int) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	key := cooldownKey{personaID, channelID}
	last, ok := dm.cooldowns[key]
	if !ok {
		return true
	}
	return time.Since(last) >= time.Duration(cooldownSecs)*time.Second
}

func (dm *DecisionMaker) resetCooldown(personaID, channelID int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.cooldowns, cooldownKey{personaID, channelID})
}
