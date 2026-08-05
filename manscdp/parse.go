package manscdp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Parse parses a MANSCDP XML body into Message.
func Parse(body []byte) (*Message, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	msg := &Message{Raw: body}
	var current *CatalogItem
	inDeviceList := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			name := localName(el.Name.Local)
			if msg.Root == "" && isRootTag(name) {
				msg.Root = name
			}
			if name == "DeviceList" {
				inDeviceList = true
			}
			if inDeviceList && name == "Item" {
				current = &CatalogItem{}
			}
		case xml.EndElement:
			name := localName(el.Name.Local)
			if name == "Item" && current != nil {
				msg.Items = append(msg.Items, *current)
				current = nil
			}
			if name == "DeviceList" {
				inDeviceList = false
			}
		case xml.CharData:
			text := strings.TrimSpace(string(el))
			if text == "" {
				continue
			}
			if current != nil {
				continue
			}
			switch {
			case strings.Contains(string(body), "<CmdType>"+text+"</CmdType>"):
				if msg.CmdType == "" {
					msg.CmdType = text
				}
			}
		}
	}

	msg.CmdType = extractTag(body, "CmdType")
	msg.SN = extractTag(body, "SN")
	msg.DeviceID = extractTag(body, "DeviceID")
	msg.Result = extractTag(body, "Result")
	msg.NotifyType = extractTag(body, "NotifyType")
	fmt.Sscanf(extractTag(body, "SumNum"), "%d", &msg.SumNum)

	if len(msg.Items) == 0 {
		msg.Items = parseCatalogItemsFallback(body)
	} else {
		nonEmpty := 0
		for _, it := range msg.Items {
			if it.DeviceID != "" {
				nonEmpty++
			}
		}
		if nonEmpty == 0 {
			msg.Items = parseCatalogItemsFallback(body)
		}
	}
	if msg.CmdType == "Catalog" && len(msg.Items) == 0 {
		msg.Items = parseCatalogItemsFallback(body)
	}
	normalizeCatalogEvents(msg.Items)
	if msg.CmdType == "RecordInfo" {
		msg.RecordItems = parseRecordItemsFallback(body)
	}
	if msg.CmdType == "PresetQuery" {
		msg.PresetItems = parsePresetItemsFallback(body)
		if msg.SumNum == 0 {
			if n := extractPresetListNum(body); n > 0 {
				msg.SumNum = n
			} else {
				msg.SumNum = len(msg.PresetItems)
			}
		}
	}
	if msg.CmdType == "Alarm" {
		msg.Alarm = parseAlarmNotify(body)
	}
	if msg.CmdType == "MobilePosition" {
		msg.Position = parseMobilePositionNotify(body)
	}
	if msg.CmdType == "DeviceStatus" {
		msg.DeviceStatus = parseDeviceStatus(body)
	}
	if msg.CmdType == "MediaStatus" {
		msg.MediaStatus = &MediaStatusNotify{
			NotifyType: msg.NotifyType,
			DeviceID:   msg.DeviceID,
		}
	}
	if msg.CmdType == "DeviceControl" {
		msg.PTZCmd = extractTag(body, "PTZCmd")
	}
	if msg.Root == "" {
		for _, tag := range []string{"Notify", "Response", "Query", "Control"} {
			if strings.Contains(string(body), "<"+tag+">") {
				msg.Root = tag
				break
			}
		}
	}
	return msg, nil
}

func isRootTag(name string) bool {
	switch name {
	case "Notify", "Response", "Query", "Control":
		return true
	default:
		return false
	}
}

func localName(name string) string {
	if idx := strings.Index(name, ":"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func extractTag(body []byte, tag string) string {
	return ExtractTag(body, tag)
}

// ExtractTag returns the text content of the first matching XML tag (case-insensitive).
func ExtractTag(body []byte, tag string) string {
	return extractTagFold(string(body), tag)
}

func extractTagFold(s, tag string) string {
	lower := strings.ToLower(s)
	tagLower := strings.ToLower(tag)
	needle := "<" + tagLower
	i := 0
	for {
		idx := strings.Index(lower[i:], needle)
		if idx < 0 {
			return ""
		}
		abs := i + idx
		after := abs + len(needle)
		if after >= len(lower) {
			return ""
		}
		ch := lower[after]
		if ch != '>' && ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' && ch != '/' {
			i = after
			continue
		}
		gt := strings.IndexByte(lower[after:], '>')
		if gt < 0 {
			return ""
		}
		contentStart := after + gt + 1
		endTag := "</" + tagLower + ">"
		j := strings.Index(lower[contentStart:], endTag)
		if j < 0 {
			return ""
		}
		return strings.TrimSpace(s[contentStart : contentStart+j])
	}
}

func parseCatalogItemsFallback(body []byte) []CatalogItem {
	s := string(body)
	lower := strings.ToLower(s)
	items := make([]CatalogItem, 0)
	for {
		i := strings.Index(lower, "<item")
		if i < 0 {
			break
		}
		closeIdx := strings.Index(lower[i:], "</item>")
		if closeIdx < 0 {
			break
		}
		end := i + closeIdx + len("</item>")
		chunk := s[i:end]
		item := CatalogItem{
			DeviceID:     extractTagFold(chunk, "DeviceID"),
			Name:         extractTagFold(chunk, "Name"),
			Manufacturer: extractTagFold(chunk, "Manufacturer"),
			Model:        extractTagFold(chunk, "Model"),
			Owner:        extractTagFold(chunk, "Owner"),
			CivilCode:    extractTagFold(chunk, "CivilCode"),
			Block:        extractTagFold(chunk, "Block"),
			Address:      extractTagFold(chunk, "Address"),
			ParentID:     extractTagFold(chunk, "ParentID"),
			Status:       extractTagFold(chunk, "Status"),
			Event:        extractTagFold(chunk, "Event"),
			OperateType:  extractTagFold(chunk, "OperateType"),
		}
		fmt.Sscanf(extractTagFold(chunk, "Parental"), "%d", &item.Parental)
		fmt.Sscanf(extractTagFold(chunk, "RegisterWay"), "%d", &item.RegisterWay)
		fmt.Sscanf(extractTagFold(chunk, "Secrecy"), "%d", &item.Secrecy)
		fmt.Sscanf(extractTagFold(chunk, "Longitude"), "%f", &item.Longitude)
		fmt.Sscanf(extractTagFold(chunk, "Latitude"), "%f", &item.Latitude)
		if info := extractInfoPTZType(chunk); info >= 0 {
			item.PTZType = info
		}
		items = append(items, item)
		s = s[end:]
		lower = strings.ToLower(s)
	}
	return items
}

func extractInfoPTZType(chunk string) int {
	i := strings.Index(strings.ToLower(chunk), "<info>")
	if i < 0 {
		return -1
	}
	sub := chunk[i:]
	val := extractTagFold(sub, "PTZType")
	var ptzType int
	fmt.Sscanf(val, "%d", &ptzType)
	return ptzType
}

var presetListNumRe = regexp.MustCompile(`(?i)<PresetList[^>]*\bNum\s*=\s*["']?(\d+)`)

func extractPresetListNum(body []byte) int {
	m := presetListNumRe.FindSubmatch(body)
	if len(m) < 2 {
		return 0
	}
	var n int
	fmt.Sscanf(string(m[1]), "%d", &n)
	return n
}

func parsePresetItemsFallback(body []byte) []Preset {
	s := string(body)
	lower := strings.ToLower(s)
	items := make([]Preset, 0)

	search := s
	searchLower := lower
	if i := strings.Index(lower, "<presetlist"); i >= 0 {
		if j := strings.Index(lower[i:], "</presetlist>"); j >= 0 {
			search = s[i : i+j]
			searchLower = strings.ToLower(search)
		}
	}

	rest := search
	restLower := searchLower
	for {
		i := strings.Index(restLower, "<item")
		if i < 0 {
			break
		}
		closeIdx := strings.Index(restLower[i:], "</item>")
		if closeIdx < 0 {
			break
		}
		end := i + closeIdx + len("</item>")
		chunk := rest[i:end]
		id := extractTagFold(chunk, "PresetID")
		name := extractTagFold(chunk, "PresetName")
		if id != "" {
			items = append(items, Preset{PresetID: id, PresetName: name})
		}
		rest = rest[end:]
		restLower = restLower[end:]
	}
	return items
}

func parseRecordItemsFallback(body []byte) []RecordItem {
	s := string(body)
	items := make([]RecordItem, 0)
	for {
		i := strings.Index(s, "<Item")
		if i < 0 {
			break
		}
		j := strings.Index(s[i:], "</Item>")
		if j < 0 {
			break
		}
		chunk := s[i : i+j+7]
		item := RecordItem{
			DeviceID:   extractTag([]byte(chunk), "DeviceID"),
			Name:       extractTag([]byte(chunk), "Name"),
			FilePath:   extractTag([]byte(chunk), "FilePath"),
			FileSize:   extractTag([]byte(chunk), "FileSize"),
			StartTime:  extractTag([]byte(chunk), "StartTime"),
			EndTime:    extractTag([]byte(chunk), "EndTime"),
			Type:       extractTag([]byte(chunk), "Type"),
			RecorderID: extractTag([]byte(chunk), "RecorderID"),
		}
		fmt.Sscanf(extractTag([]byte(chunk), "Secrecy"), "%d", &item.Secrecy)
		items = append(items, item)
		s = s[i+j+7:]
	}
	return items
}

func parseAlarmNotify(body []byte) *AlarmNotify {
	a := &AlarmNotify{
		AlarmPriority:    extractTag(body, "AlarmPriority"),
		AlarmMethod:      extractTag(body, "AlarmMethod"),
		AlarmTime:        extractTag(body, "AlarmTime"),
		AlarmDescription: extractTag(body, "AlarmDescription"),
	}
	fmt.Sscanf(extractTag(body, "Longitude"), "%f", &a.Longitude)
	fmt.Sscanf(extractTag(body, "Latitude"), "%f", &a.Latitude)
	fmt.Sscanf(extractTag(body, "AlarmType"), "%d", &a.AlarmType)
	return a
}

func parseMobilePositionNotify(body []byte) *MobilePositionNotify {
	p := &MobilePositionNotify{Time: extractTag(body, "Time")}
	fmt.Sscanf(extractTag(body, "Longitude"), "%f", &p.Longitude)
	fmt.Sscanf(extractTag(body, "Latitude"), "%f", &p.Latitude)
	fmt.Sscanf(extractTag(body, "Speed"), "%f", &p.Speed)
	fmt.Sscanf(extractTag(body, "Direction"), "%f", &p.Direction)
	fmt.Sscanf(extractTag(body, "Altitude"), "%f", &p.Altitude)
	return p
}

func parseDeviceStatus(body []byte) *DeviceStatus {
	return &DeviceStatus{
		Result:      extractTag(body, "Result"),
		Online:      extractTag(body, "Online"),
		Status:      extractTag(body, "Status"),
		Encode:      extractTag(body, "Encode"),
		Record:      extractTag(body, "Record"),
		DeviceTime:  extractTag(body, "DeviceTime"),
		AlarmStatus: extractTag(body, "Alarmstatus"),
	}
}

// normalizeCatalogEvents fills Event from OperateType when Event is empty (Hik-style).
func normalizeCatalogEvents(items []CatalogItem) {
	for i := range items {
		if items[i].Event == "" && items[i].OperateType != "" {
			items[i].Event = items[i].OperateType
		}
	}
}
