package manscdp

import (
	"fmt"
	"strings"
)

// BuildCatalogQuery builds a Catalog query MESSAGE body.
func BuildCatalogQuery(deviceID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>Catalog</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, deviceID)
}

// BuildDeviceInfoQuery builds a DeviceInfo query MESSAGE body.
func BuildDeviceInfoQuery(deviceID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>DeviceInfo</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, deviceID)
}

// BuildDeviceStatusQuery builds a DeviceStatus query MESSAGE body.
func BuildDeviceStatusQuery(deviceID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>DeviceStatus</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, deviceID)
}

// BuildDeviceControlPTZ builds a DeviceControl body with PTZCmd.
func BuildDeviceControlPTZ(channelID, sn, ptzCmd string) string {
	if sn == "" {
		sn = "1"
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<PTZCmd>%s</PTZCmd>
</Control>`, sn, channelID, ptzCmd)
}

// BuildPresetQuery builds a PresetQuery MESSAGE body.
func BuildPresetQuery(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>PresetQuery</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, channelID)
}

// BuildDeviceControlFrontEnd builds a DeviceControl body for front-end commands.
func BuildDeviceControlFrontEnd(channelID, sn, ptzCmd string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<PTZCmd>%s</PTZCmd>
</Control>`, sn, channelID, ptzCmd)
}

// BuildRecordCmd builds DeviceControl with RecordCmd (Record / StopRecord).
func BuildRecordCmd(channelID, sn, recordCmd string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<RecordCmd>%s</RecordCmd>
</Control>`, sn, channelID, recordCmd)
}

// BuildGuardCmd builds DeviceControl with GuardCmd (SetGuard / ResetGuard).
func BuildGuardCmd(channelID, sn, guardCmd string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<GuardCmd>%s</GuardCmd>
</Control>`, sn, channelID, guardCmd)
}

// BuildTeleBoot builds DeviceControl TeleBoot.
func BuildTeleBoot(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<TeleBoot>Boot</TeleBoot>
</Control>`, sn, channelID)
}

// BuildIFrameCmd builds DeviceControl IFrameCmd (强制关键帧).
func BuildIFrameCmd(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<IFrameCmd>Send</IFrameCmd>
</Control>`, sn, channelID)
}

// BuildDragZoom builds DeviceControl DragZoomIn / DragZoomOut.
func BuildDragZoom(channelID, sn string, zoomIn bool, length, width, midX, midY, lengthX, lengthY int) string {
	tag := "DragZoomOut"
	if zoomIn {
		tag = "DragZoomIn"
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<%s>
<Length>%d</Length>
<Width>%d</Width>
<MidPointX>%d</MidPointX>
<MidPointY>%d</MidPointY>
<LengthX>%d</LengthX>
<LengthY>%d</LengthY>
</%s>
</Control>`, sn, channelID, tag, length, width, midX, midY, lengthX, lengthY, tag)
}

// BuildHomePosition builds DeviceControl HomePosition.
func BuildHomePosition(channelID, sn string, enabled, resetTime, presetIndex int) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<HomePosition>
<Enabled>%d</Enabled>
<ResetTime>%d</ResetTime>
<PresetIndex>%d</PresetIndex>
</HomePosition>
</Control>`, sn, channelID, enabled, resetTime, presetIndex)
}

// BuildBasicParamConfig builds DeviceConfig BasicParam.
func BuildBasicParamConfig(deviceID, sn, name string, expiration, heartBeatInterval, heartBeatCount, positionCapability int) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceConfig</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<BasicParam>
<Name>%s</Name>
<Expiration>%d</Expiration>
<HeartBeatInterval>%d</HeartBeatInterval>
<HeartBeatCount>%d</HeartBeatCount>
<PositionCapability>%d</PositionCapability>
</BasicParam>
</Control>`, sn, deviceID, name, expiration, heartBeatInterval, heartBeatCount, positionCapability)
}

// BuildAlarmCmd builds DeviceControl AlarmCmd (e.g. ResetAlarm).
func BuildAlarmCmd(channelID, sn, alarmCmd string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<AlarmCmd>%s</AlarmCmd>
</Control>`, sn, channelID, alarmCmd)
}

// BuildRecordInfoQuery builds a RecordInfo query (Type=all, Secrecy=0).
func BuildRecordInfoQuery(channelID, sn, startTime, endTime string) string {
	return BuildRecordInfoQueryOpts(channelID, sn, RecordInfoOpts{
		StartTime: startTime,
		EndTime:   endTime,
		Secrecy:   0,
		Type:      "all",
	})
}

// BuildRecordInfoQueryOpts builds a RecordInfo query with optional RecLocation/RecordPos.
func BuildRecordInfoQueryOpts(channelID, sn string, opts RecordInfoOpts) string {
	typ := opts.Type
	if typ == "" {
		typ = "all"
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="GB2312"?>` + "\n")
	b.WriteString("<Query>\n")
	b.WriteString("<CmdType>RecordInfo</CmdType>\n")
	b.WriteString(fmt.Sprintf("<SN>%s</SN>\n", sn))
	b.WriteString(fmt.Sprintf("<DeviceID>%s</DeviceID>\n", channelID))
	b.WriteString(fmt.Sprintf("<StartTime>%s</StartTime>\n", opts.StartTime))
	b.WriteString(fmt.Sprintf("<EndTime>%s</EndTime>\n", opts.EndTime))
	b.WriteString(fmt.Sprintf("<Secrecy>%d</Secrecy>\n", opts.Secrecy))
	b.WriteString(fmt.Sprintf("<Type>%s</Type>\n", typ))
	if opts.RecLocation != "" {
		b.WriteString(fmt.Sprintf("<RecLocation>%s</RecLocation>\n", opts.RecLocation))
	}
	if opts.RecordPos != "" {
		b.WriteString(fmt.Sprintf("<RecordPos>%s</RecordPos>\n", opts.RecordPos))
	}
	b.WriteString("</Query>")
	return b.String()
}

// BuildConfigDownloadQuery builds a ConfigDownload query.
func BuildConfigDownloadQuery(deviceID, sn, configType string) string {
	if configType == "" {
		configType = "BasicParam"
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>ConfigDownload</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<ConfigType>%s</ConfigType>
</Query>`, sn, deviceID, configType)
}

// BuildHomePositionQuery builds a HomePositionQuery request.
func BuildHomePositionQuery(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>HomePositionQuery</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, channelID)
}

// BuildCruiseTrackListQuery builds a CruiseTrackListQuery request.
func BuildCruiseTrackListQuery(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>CruiseTrackListQuery</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, channelID)
}

// BuildCruiseTrackQuery builds a CruiseTrackQuery request for a track number.
func BuildCruiseTrackQuery(channelID, sn string, number int) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>CruiseTrackQuery</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<Number>%d</Number>
</Query>`, sn, channelID, number)
}

// BuildPTZPositionQuery builds a PTZPosition query/subscribe body.
func BuildPTZPositionQuery(channelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>PTZPosition</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, channelID)
}

// BuildPlatformKeepalive builds an upstream platform keepalive notify.
func BuildPlatformKeepalive(platformID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Notify>
<CmdType>Keepalive</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<Status>OK</Status>
</Notify>`, sn, platformID)
}

// BuildCatalogNotify builds a Catalog notify for cascade.
func BuildCatalogNotify(platformDeviceID, sn string, items []CatalogItem) string {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="GB2312"?>` + "\r\n")
	buf.WriteString("<Notify>\r\n")
	buf.WriteString("<CmdType>Catalog</CmdType>\r\n")
	buf.WriteString(fmt.Sprintf("<SN>%s</SN>\r\n", sn))
	buf.WriteString(fmt.Sprintf("<DeviceID>%s</DeviceID>\r\n", platformDeviceID))
	buf.WriteString(fmt.Sprintf("<SumNum>%d</SumNum>\r\n", len(items)))
	buf.WriteString(fmt.Sprintf("<DeviceList Num=\"%d\">\r\n", len(items)))
	for _, it := range items {
		buf.WriteString("<Item>\r\n")
		buf.WriteString(fmt.Sprintf("<DeviceID>%s</DeviceID>\r\n", it.DeviceID))
		buf.WriteString(fmt.Sprintf("<Name>%s</Name>\r\n", it.Name))
		if it.Status != "" {
			buf.WriteString(fmt.Sprintf("<Status>%s</Status>\r\n", it.Status))
		}
		ev := it.Event
		if ev == "" {
			ev = it.OperateType
		}
		if ev != "" {
			buf.WriteString(fmt.Sprintf("<Event>%s</Event>\r\n", ev))
		}
		buf.WriteString("</Item>\r\n")
	}
	buf.WriteString("</DeviceList>\r\n</Notify>\r\n")
	return buf.String()
}

// BuildBroadcastNotify builds an audio broadcast notify.
func BuildBroadcastNotify(sourceID, targetChannelID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>`+"\r\n"+
		"<Notify>\r\n"+
		"<CmdType>Broadcast</CmdType>\r\n"+
		"<SN>%s</SN>\r\n"+
		"<SourceID>%s</SourceID>\r\n"+
		"<TargetID>%s</TargetID>\r\n"+
		"</Notify>\r\n", sn, sourceID, targetChannelID)
}

// BuildSubscribeCatalog builds a Catalog subscribe query body.
func BuildSubscribeCatalog(deviceID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>Catalog</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, deviceID)
}

// BuildSubscribeAlarm builds an Alarm subscribe query body.
func BuildSubscribeAlarm(deviceID, sn string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>Alarm</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
</Query>`, sn, deviceID)
}

// BuildSubscribeMobilePosition builds a MobilePosition subscribe query body.
func BuildSubscribeMobilePosition(deviceID, sn string, interval int) string {
	if interval <= 0 {
		interval = 5
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>MobilePosition</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<Interval>%d</Interval>
</Query>`, sn, deviceID, interval)
}
