package manscdp

import "testing"

func TestParsePresetItems(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<Response>
<CmdType>PresetQuery</CmdType>
<SN>12</SN>
<DeviceID>34020000001320000001</DeviceID>
<SumNum>2</SumNum>
<PresetList Num="2">
<Item>
<PresetID>1</PresetID>
<PresetName>门口</PresetName>
</Item>
<Item>
<PresetID>2</PresetID>
<PresetName>大厅</PresetName>
</Item>
</PresetList>
</Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if msg.CmdType != "PresetQuery" {
		t.Fatalf("CmdType=%s", msg.CmdType)
	}
	if msg.SumNum != 2 {
		t.Fatalf("SumNum=%d", msg.SumNum)
	}
	if len(msg.PresetItems) != 2 {
		t.Fatalf("items=%d", len(msg.PresetItems))
	}
	if msg.PresetItems[0].PresetID != "1" || msg.PresetItems[0].PresetName != "门口" {
		t.Fatalf("item0=%+v", msg.PresetItems[0])
	}
}

func TestParsePresetEmptyList(t *testing.T) {
	body := []byte(`<Response>
<CmdType>PresetQuery</CmdType>
<SN>1</SN>
<DeviceID>34020000001320000001</DeviceID>
<SumNum>0</SumNum>
<PresetList Num="0">
</PresetList>
</Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.PresetItems) != 0 {
		t.Fatalf("expected empty, got %+v", msg.PresetItems)
	}
}

func TestParseCatalogOperateTypeAsEvent(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<Notify>
<CmdType>Catalog</CmdType>
<SN>1</SN>
<DeviceID>130909113319427420</DeviceID>
<SumNum>1</SumNum>
<DeviceList Num="1">
<Item>
<DeviceID>130909113319427421</DeviceID>
<Name>cam</Name>
<Status>OFF</Status>
<OperateType>ADD</OperateType>
</Item>
</DeviceList>
</Notify>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.Items) != 1 {
		t.Fatalf("items=%d", len(msg.Items))
	}
	if msg.Items[0].Event != "ADD" {
		t.Fatalf("Event=%q OperateType=%q", msg.Items[0].Event, msg.Items[0].OperateType)
	}
}

func TestParseDeviceStatus(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<Response>
<CmdType>DeviceStatus</CmdType>
<SN>9</SN>
<DeviceID>130909113319427420</DeviceID>
<Result>OK</Result>
<Online>OFFLINE</Online>
<Status>OK</Status>
<Encode>ON</Encode>
<Record>ON</Record>
<DeviceTime>2013-09-10T12:00:00</DeviceTime>
</Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if msg.DeviceStatus == nil || msg.DeviceStatus.Online != "OFFLINE" || msg.DeviceStatus.Record != "ON" {
		t.Fatalf("status=%+v", msg.DeviceStatus)
	}
}

func TestParseMediaStatus(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<Notify>
<CmdType>MediaStatus</CmdType>
<SN>1</SN>
<DeviceID>130909113319427420</DeviceID>
<NotifyType>121</NotifyType>
</Notify>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if msg.MediaStatus == nil || msg.MediaStatus.NotifyType != "121" {
		t.Fatalf("media=%+v", msg.MediaStatus)
	}
}

func TestBuildRecordInfoOpts(t *testing.T) {
	body := BuildRecordInfoQueryOpts("ch1", "2", RecordInfoOpts{
		StartTime: "2013-01-01 00:00:00", EndTime: "2013-01-02 00:00:00",
		Type: "all", RecLocation: "2", RecordPos: "2",
	})
	if !containsAll(body, "<RecLocation>2</RecLocation>", "<RecordPos>2</RecordPos>") {
		t.Fatalf("body=%s", body)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !containsStr(s, p) {
			return false
		}
	}
	return true
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
