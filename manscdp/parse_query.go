package manscdp

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// This file parses the query responses that carry nested structures
// (ConfigDownload, and the 2022 query family). encoding/xml is used instead of
// the tag scanner because these payloads nest lists inside items.

type basicParamXML struct {
	Result     string `xml:"Result"`
	BasicParam struct {
		Name               string `xml:"Name"`
		Expiration         string `xml:"Expiration"`
		HeartBeatInterval  string `xml:"HeartBeatInterval"`
		HeartBeatCount     string `xml:"HeartBeatCount"`
		PositionCapability string `xml:"PositionCapability"`
		Longitude          string `xml:"Longitude"`
		Latitude           string `xml:"Latitude"`
	} `xml:"BasicParam"`
}

func parseBasicParam(body []byte) *BasicParam {
	var doc basicParamXML
	if err := xml.Unmarshal(body, &doc); err != nil {
		return &BasicParam{Result: ExtractTag(body, "Result")}
	}
	return &BasicParam{
		Result:             valueOr(doc.Result, ExtractTag(body, "Result")),
		Name:               doc.BasicParam.Name,
		Expiration:         doc.BasicParam.Expiration,
		HeartBeatInterval:  doc.BasicParam.HeartBeatInterval,
		HeartBeatCount:     doc.BasicParam.HeartBeatCount,
		PositionCapability: doc.BasicParam.PositionCapability,
		Longitude:          doc.BasicParam.Longitude,
		Latitude:           doc.BasicParam.Latitude,
	}
}

type homePositionXML struct {
	Result       string `xml:"Result"`
	HomePosition struct {
		Enabled     int `xml:"Enabled"`
		ResetTime   int `xml:"ResetTime"`
		PresetIndex int `xml:"PresetIndex"`
	} `xml:"HomePosition"`
}

func parseHomePosition(body []byte) *HomePosition {
	var doc homePositionXML
	_ = xml.Unmarshal(body, &doc)
	hp := &HomePosition{
		Result:      valueOr(doc.Result, ExtractTag(body, "Result")),
		Enabled:     doc.HomePosition.Enabled,
		ResetTime:   doc.HomePosition.ResetTime,
		PresetIndex: doc.HomePosition.PresetIndex,
	}
	if hp.Enabled == 0 {
		hp.Enabled = intTag(body, "Enabled")
	}
	if hp.ResetTime == 0 {
		hp.ResetTime = intTag(body, "ResetTime")
	}
	if hp.PresetIndex == 0 {
		hp.PresetIndex = intTag(body, "PresetIndex")
	}
	return hp
}

type cruiseTrackListXML struct {
	Result string `xml:"Result"`
	Num    int    `xml:"Num"`
	Tracks []struct {
		Number int    `xml:"Number"`
		Name   string `xml:"Name"`
		Points []struct {
			PointID  int    `xml:"PointID"`
			PresetID int    `xml:"PresetID"`
			Name     string `xml:"Name"`
			StayTime int    `xml:"StayTime"`
			Speed    int    `xml:"Speed"`
		} `xml:"CruisePointList>Item"`
	} `xml:"CruiseTrackList>Item"`
}

func parseCruiseTrackList(body []byte) *CruiseTrackList {
	var doc cruiseTrackListXML
	_ = xml.Unmarshal(body, &doc)
	out := &CruiseTrackList{Result: valueOr(doc.Result, ExtractTag(body, "Result")), Num: doc.Num}
	for _, t := range doc.Tracks {
		track := CruiseTrack{Number: t.Number, Name: t.Name}
		for _, p := range t.Points {
			track.Points = append(track.Points, CruisePoint{
				PointID: p.PointID, PresetID: p.PresetID, Name: p.Name, StayTime: p.StayTime, Speed: p.Speed,
			})
		}
		out.Tracks = append(out.Tracks, track)
	}
	if out.Num == 0 {
		out.Num = len(out.Tracks)
	}
	return out
}

type cruiseTrackXML struct {
	Result      string `xml:"Result"`
	CruiseTrack struct {
		Number int    `xml:"Number"`
		Name   string `xml:"Name"`
		Points []struct {
			PointID  int    `xml:"PointID"`
			PresetID int    `xml:"PresetID"`
			Name     string `xml:"Name"`
			StayTime int    `xml:"StayTime"`
			Speed    int    `xml:"Speed"`
		} `xml:"CruisePointList>Item"`
	} `xml:"CruiseTrack"`
}

func parseCruiseTrack(body []byte) *CruiseTrack {
	var doc cruiseTrackXML
	_ = xml.Unmarshal(body, &doc)
	track := &CruiseTrack{
		Number: doc.CruiseTrack.Number,
		Name:   doc.CruiseTrack.Name,
	}
	for _, p := range doc.CruiseTrack.Points {
		track.Points = append(track.Points, CruisePoint{
			PointID: p.PointID, PresetID: p.PresetID, Name: p.Name, StayTime: p.StayTime, Speed: p.Speed,
		})
	}
	return track
}

type ptzPositionXML struct {
	Result      string `xml:"Result"`
	PTZPosition struct {
		Pan  int `xml:"Pan"`
		Tilt int `xml:"Tilt"`
		Zoom int `xml:"Zoom"`
	} `xml:"PTZPosition"`
}

func parsePTZPosition(body []byte) *PTZPosition {
	var doc ptzPositionXML
	_ = xml.Unmarshal(body, &doc)
	pos := &PTZPosition{
		Result: valueOr(doc.Result, ExtractTag(body, "Result")),
		Pan:    doc.PTZPosition.Pan,
		Tilt:   doc.PTZPosition.Tilt,
		Zoom:   doc.PTZPosition.Zoom,
	}
	// 部分实现使用 PanValue/TiltValue/ZoomValue。
	if pos.Pan == 0 {
		pos.Pan = intTag(body, "PanValue")
	}
	if pos.Tilt == 0 {
		pos.Tilt = intTag(body, "TiltValue")
	}
	if pos.Zoom == 0 {
		pos.Zoom = intTag(body, "ZoomValue")
	}
	return pos
}

type sdCardStatusXML struct {
	Result       string `xml:"Result"`
	SDCardStatus struct {
		Status       string `xml:"Status"`
		Type         string `xml:"Type"`
		Capacity     int    `xml:"Capacity"`
		FreeCapacity int    `xml:"FreeCapacity"`
	} `xml:"SDCardStatus"`
}

func parseSDCardStatus(body []byte) *SDCardStatus {
	var doc sdCardStatusXML
	_ = xml.Unmarshal(body, &doc)
	out := &SDCardStatus{
		Result:     valueOr(doc.Result, ExtractTag(body, "Result")),
		Status:     doc.SDCardStatus.Status,
		Type:       doc.SDCardStatus.Type,
		CapacityMB: doc.SDCardStatus.Capacity,
		FreeMB:     doc.SDCardStatus.FreeCapacity,
	}
	if out.Status == "" {
		out.Status = ExtractTag(body, "Status")
	}
	return out
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func intTag(body []byte, tag string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(ExtractTag(body, tag)))
	return v
}
