package ssrc

import "testing"

func TestPlay(t *testing.T) {
	got := Play("3402000000", 1)
	want := "0200000001"
	if got != want {
		t.Fatalf("Play() = %q, want %q", got, want)
	}
}
