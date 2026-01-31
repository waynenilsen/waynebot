package tools

import "slices"

// SandboxConfig defines security boundaries for tool execution.
type SandboxConfig struct {
	AllowedCommands []string
	BaseDir         string
	BlockedHosts    []string
}

// DefaultBlockedHosts returns the standard set of hosts that should be blocked
// to prevent SSRF attacks.
func DefaultBlockedHosts() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"169.254.169.254",
		"::1",
	}
}

// IsCommandAllowed checks whether a command is in the allowed list.
func (s *SandboxConfig) IsCommandAllowed(cmd string) bool {
	return slices.Contains(s.AllowedCommands, cmd)
}

// IsHostBlocked checks whether a host is in the blocked list.
func (s *SandboxConfig) IsHostBlocked(host string) bool {
	return slices.Contains(s.BlockedHosts, host)
}
