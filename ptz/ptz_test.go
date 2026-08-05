package ptz

import "testing"

func TestFrontEndCmdPreset(t *testing.T) {
	got := FrontEndCmd(0x81, 0, 1, 0)
	wantPrefix := "A50F01810001"
	if got[:12] != wantPrefix {
		t.Fatalf("set preset cmd prefix = %s, want %s...", got, wantPrefix)
	}
	call := FrontEndCmd(0x82, 0, 5, 0)
	if call[:12] != "A50F01820005" {
		t.Fatalf("call preset cmd = %s", call)
	}
	del := FrontEndCmd(0x83, 0, 5, 0)
	if del[:12] != "A50F01830005" {
		t.Fatalf("delete preset cmd = %s", del)
	}
}
