package sigsci

import "testing"

func TestStripPort(t *testing.T) {
	got := StripPort("127.0.0.1:8000")
	if got != "127.0.0.1" {
		t.Errorf("StripPort(%q) = %q, want %q", "127.0.0.1:8000", got, "127.0.0.1")
	}
}
