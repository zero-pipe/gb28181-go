package manscdp

import "testing"

func TestParseConfigDownloadBasicParam(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Response>
<CmdType>ConfigDownload</CmdType>
<SN>7</SN>
<DeviceID>34020000001320000001</DeviceID>
<Result>OK</Result>
<BasicParam>
<Name>Camera-1</Name>
<Expiration>3600</Expiration>
<HeartBeatInterval>60</HeartBeatInterval>
<HeartBeatCount>3</HeartBeatCount>
<PositionCapability>1</PositionCapability>
</BasicParam>
</Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.BasicParam == nil {
		t.Fatal("BasicParam not parsed")
	}
	if msg.BasicParam.Name != "Camera-1" || msg.BasicParam.HeartBeatCount != "3" {
		t.Fatalf("unexpected BasicParam: %+v", *msg.BasicParam)
	}
}

func TestParsePTZPosition(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Response><CmdType>PTZPosition</CmdType><SN>1</SN>
<DeviceID>ch1</DeviceID><Result>OK</Result>
<PTZPosition><Pan>120</Pan><Tilt>-30</Tilt><Zoom>4</Zoom></PTZPosition></Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.PTZPosition == nil || msg.PTZPosition.Pan != 120 || msg.PTZPosition.Tilt != -30 || msg.PTZPosition.Zoom != 4 {
		t.Fatalf("unexpected PTZPosition: %+v", msg.PTZPosition)
	}
}

func TestParseCruiseTrackList(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Response><CmdType>CruiseTrackListQuery</CmdType><SN>1</SN>
<DeviceID>ch1</DeviceID><Result>OK</Result><Num>1</Num>
<CruiseTrackList Num="1">
  <Item>
    <Number>1</Number><Name>track-a</Name>
    <CruisePointList Num="2">
      <Item><PointID>1</PointID><PresetID>2</PresetID><StayTime>5</StayTime></Item>
      <Item><PointID>2</PointID><PresetID>3</PresetID><StayTime>8</StayTime></Item>
    </CruisePointList>
  </Item>
</CruiseTrackList></Response>`)
	msg, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	list := msg.CruiseTrackList
	if list == nil || len(list.Tracks) != 1 {
		t.Fatalf("unexpected list: %+v", list)
	}
	if list.Tracks[0].Name != "track-a" || len(list.Tracks[0].Points) != 2 {
		t.Fatalf("unexpected track: %+v", list.Tracks[0])
	}
	if list.Tracks[0].Points[1].PresetID != 3 || list.Tracks[0].Points[1].StayTime != 8 {
		t.Fatalf("unexpected point: %+v", list.Tracks[0].Points[1])
	}
}
