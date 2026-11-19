package manscdp

import (
	"strings"
	"testing"
)

func TestBuildPTZPreciseCtrl(t *testing.T) {
	out := BuildPTZPreciseCtrl("ch1", "5", 120.5, -10, 4)
	for _, want := range []string{"<PTZPreciseCtrlType>", "<Pan>120.5</Pan>", "<Tilt>-10</Tilt>", "<Zoom>4</Zoom>", "<SN>5</SN>"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestBuildFormatAndTargetTrack(t *testing.T) {
	if out := BuildFormatSDCard("ch1", "1"); !strings.Contains(out, "<FormatSDCard>") {
		t.Fatalf("format cmd: %s", out)
	}
	out := BuildTargetTrack("ch1", "1", "Stop")
	if !strings.Contains(out, "<TargetTrackType>Stop</TargetTrackType>") {
		t.Fatalf("target track: %s", out)
	}
}
