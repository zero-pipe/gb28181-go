package sdp

import "testing"

func TestParseOfferPlayback(t *testing.T) {
	body := []byte("v=0\r\no=34020000001320000001 0 0 IN IP4 10.0.0.1\r\ns=Playback\r\nu=ch1:0\r\nc=IN IP4 10.0.0.1\r\nt=1700000000 1700003600\r\nm=video 10000 RTP/AVP 96\r\ny=0123456789\r\n")
	info := ParseOffer(body)
	if info.Kind != OfferPlayback {
		t.Fatalf("kind=%q", info.Kind)
	}
	if info.StartTime == "" || info.EndTime == "" {
		t.Fatalf("missing time range: %+v", info)
	}
	if info.SSRC != "0123456789" {
		t.Fatalf("ssrc=%q", info.SSRC)
	}
}

func TestParseOfferDownload(t *testing.T) {
	info := ParseOffer([]byte("v=0\ns=Download\nt=1 2\n"))
	if info.Kind != OfferDownload {
		t.Fatalf("kind=%q", info.Kind)
	}
}

func TestParseOfferDefaultsToPlay(t *testing.T) {
	info := ParseOffer([]byte("v=0\ns=Play\nt=0 0\n"))
	if info.Kind != OfferPlay {
		t.Fatalf("kind=%q", info.Kind)
	}
}
