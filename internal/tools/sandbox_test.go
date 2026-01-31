package tools

import "testing"

func TestDefaultBlockedHosts(t *testing.T) {
	hosts := DefaultBlockedHosts()
	want := map[string]bool{
		"localhost":       true,
		"127.0.0.1":       true,
		"0.0.0.0":         true,
		"169.254.169.254": true,
		"::1":             true,
	}
	if len(hosts) != len(want) {
		t.Fatalf("got %d hosts, want %d", len(hosts), len(want))
	}
	for _, h := range hosts {
		if !want[h] {
			t.Errorf("unexpected host %q", h)
		}
	}
}

func TestIsCommandAllowed(t *testing.T) {
	cfg := SandboxConfig{AllowedCommands: []string{"ls", "cat", "grep"}}

	if !cfg.IsCommandAllowed("ls") {
		t.Error("ls should be allowed")
	}
	if cfg.IsCommandAllowed("rm") {
		t.Error("rm should not be allowed")
	}
}

func TestIsHostBlocked(t *testing.T) {
	cfg := SandboxConfig{BlockedHosts: DefaultBlockedHosts()}

	if !cfg.IsHostBlocked("localhost") {
		t.Error("localhost should be blocked")
	}
	if cfg.IsHostBlocked("example.com") {
		t.Error("example.com should not be blocked")
	}
}
