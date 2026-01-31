package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration, loaded from environment variables.
type Config struct {
	Port       int
	DBPath     string
	SandboxDir string

	CORSOrigins []string

	IMAPHost    string
	IMAPPort    int
	IMAPUser    string
	IMAPPass    string
	IMAPChannel string

	OpenRouterKey string

	ArchiveDir string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	c := Config{
		Port:       envInt("WAYNEBOT_PORT", 59731),
		DBPath:     envStr("WAYNEBOT_DB_PATH", "waynebot.db"),
		SandboxDir: envStr("WAYNEBOT_SANDBOX_DIR", "/tmp/waynebot-sandbox"),

		CORSOrigins: envList("WAYNEBOT_CORS_ORIGINS", []string{"http://localhost:5173"}),

		IMAPHost:    envStr("WAYNEBOT_IMAP_HOST", ""),
		IMAPPort:    envInt("WAYNEBOT_IMAP_PORT", 993),
		IMAPUser:    envStr("WAYNEBOT_IMAP_USER", ""),
		IMAPPass:    envStr("WAYNEBOT_IMAP_PASS", ""),
		IMAPChannel: envStr("WAYNEBOT_IMAP_CHANNEL", "email"),

		OpenRouterKey: envStr("WAYNEBOT_OPENROUTER_KEY", ""),

		ArchiveDir: envStr("WAYNEBOT_ARCHIVE_DIR", "./archives"),
	}
	return c
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return fallback
}
