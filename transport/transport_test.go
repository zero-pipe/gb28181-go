package transport

import "testing"

func TestNormalize(t *testing.T) {
	if got := Normalize(""); got != UDP {
		t.Fatalf("empty -> %q", got)
	}
	if got := Normalize("TCP-PASSIVE"); got != TCPPassive {
		t.Fatalf("TCP-PASSIVE -> %q", got)
	}
	if got := Normalize("udp"); got != UDP {
		t.Fatalf("udp -> %q", got)
	}
}
