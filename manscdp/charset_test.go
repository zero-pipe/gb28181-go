package manscdp

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestParseDecodesGB18030(t *testing.T) {
	utf8xml := `<?xml version="1.0" encoding="GB2312"?>
<Notify><CmdType>Catalog</CmdType><SN>1</SN><DeviceID>34020000001110000001</DeviceID>
<DeviceList><Item><DeviceID>34020000001310000001</DeviceID><Name>大门摄像机</Name></Item></DeviceList></Notify>`
	gb, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(utf8xml))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}

	msg, err := Parse(gb)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(msg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(msg.Items))
	}
	if msg.Items[0].Name != "大门摄像机" {
		t.Fatalf("GB18030 name not decoded, got %q", msg.Items[0].Name)
	}
}

func TestEncodeBodyEmitsGB18030DeclarationAndBytes(t *testing.T) {
	body := `<?xml version="1.0" encoding="GB2312"?>
<Response><Name>前端摄像机</Name></Response>`
	out := EncodeBody(body)

	if !bytes.Contains(out, []byte(`encoding="GB18030"`)) {
		t.Fatalf("declaration not rewritten: %s", out)
	}
	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(out)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Contains(decoded, []byte("前端摄像机")) {
		t.Fatalf("body not encoded as GB18030: %s", decoded)
	}
}
