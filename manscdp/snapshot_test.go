package manscdp

import "testing"

func TestSnapShotFileNameRoundTrip(t *testing.T) {
	name := SnapShotFileName("34020000001320000001", "20260101120000123", 7) + ".jpg"
	if len(name) != 45 {
		t.Fatalf("unexpected length %d (%q)", len(name), name)
	}
	dev, ts, seq, ok := ParseSnapShotFileName(name)
	if !ok || dev != "34020000001320000001" || ts != "20260101120000123" || seq != 7 {
		t.Fatalf("parse failed: %q %q %d %v", dev, ts, seq, ok)
	}
}

func TestValidSnapShotSessionID(t *testing.T) {
	if !ValidSnapShotSessionID("snap-123") {
		t.Fatal("expected valid")
	}
	if ValidSnapShotSessionID("ab") || ValidSnapShotSessionID("bad id") {
		t.Fatal("expected invalid")
	}
}

func TestParseSnapShotFinished(t *testing.T) {
	body := []byte(`<Notify><CmdType>UploadSnapShotFinished</CmdType><SN>1</SN>
<DeviceID>34020000001320000001</DeviceID><SessionID>s-1</SessionID>
<ImageList Num="1"><Image><ImageID>a</ImageID><FilePath>http://h/1.jpg</FilePath></Image></ImageList></Notify>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.SnapShot == nil || msg.SnapShot.SessionID != "s-1" || len(msg.SnapShot.Images) != 1 {
		t.Fatalf("unexpected: %+v", msg.SnapShot)
	}
	if msg.SnapShot.Images[0].FilePath != "http://h/1.jpg" {
		t.Fatalf("unexpected image: %+v", msg.SnapShot.Images[0])
	}
}
