package embedding

import "testing"

func TestConstants(t *testing.T) {
	if Model == "" {
		t.Error("Model constant is empty")
	}
	if Dims != 1536 {
		t.Errorf("Dims = %d, want 1536", Dims)
	}
}
