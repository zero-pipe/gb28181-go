package manscdp

import (
	"fmt"
	"regexp"
	"strings"
)

// BuildSnapShotConfig builds an image-capture configuration command (9.14,
// A.2.3.2.12 snapShotCfgType). snapNum is capped at 10, interval minimum 1s.
func BuildSnapShotConfig(deviceID, sn string, snapNum, interval int, uploadURL, sessionID string) string {
	if snapNum <= 0 || snapNum > 10 {
		snapNum = 1
	}
	if interval < 1 {
		interval = 1
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceConfig</CmdType>
<SN>%s</SN>
<DeviceID>%s</DeviceID>
<SnapShotCfg>
<SnapNum>%d</SnapNum>
<Interval>%d</Interval>
<UploadURL>%s</UploadURL>
<SessionID>%s</SessionID>
</SnapShotCfg>
</Control>`, sn, deviceID, snapNum, interval, uploadURL, sessionID)
}

var snapSessionRe = regexp.MustCompile(`^[A-Za-z0-9-]{3,128}$`)

// ValidSnapShotSessionID reports whether id meets the 2022 SessionID rule
// (letters/digits/hyphen, 3–128 bytes).
func ValidSnapShotSessionID(id string) bool {
	return snapSessionRe.MatchString(id)
}

// SnapShotFileName builds the 41-character image file name:
// deviceID(20) + 02 + YYYYMMDDhhmmssSSS(17) + seq(2).
func SnapShotFileName(deviceID, timestamp string, seq int) string {
	return fmt.Sprintf("%s02%s%02d", deviceID, timestamp, seq%100)
}

// ParseSnapShotFileName parses the 41-character image file name.
func ParseSnapShotFileName(name string) (deviceID, timestamp string, seq int, ok bool) {
	base := strings.TrimSuffix(name, ".jpg")
	if len(base) != 41 {
		return "", "", 0, false
	}
	deviceID = base[:20]
	ts := base[22:39]
	if _, err := fmt.Sscanf(base[39:], "%02d", &seq); err != nil {
		return "", "", 0, false
	}
	return deviceID, ts, seq, true
}

func parseSnapShotFinished(body []byte) *SnapShotFinished {
	out := &SnapShotFinished{SessionID: ExtractTag(body, "SessionID")}
	for _, block := range extractTagBlocks(string(body), "Image") {
		img := SnapShotImage{
			ImageID:  strings.TrimSpace(extractTagFold(block, "ImageID")),
			FilePath: strings.TrimSpace(extractTagFold(block, "FilePath")),
		}
		if img.ImageID == "" && img.FilePath == "" {
			continue
		}
		out.Images = append(out.Images, img)
	}
	return out
}

// extractTagBlocks returns the inner text of every <tag>...</tag> occurrence.
func extractTagBlocks(s, tag string) []string {
	open := "<" + strings.ToLower(tag) + ">"
	closeTag := "</" + strings.ToLower(tag) + ">"
	lower := strings.ToLower(s)
	var out []string
	idx := 0
	for {
		i := strings.Index(lower[idx:], open)
		if i < 0 {
			break
		}
		start := idx + i + len(open)
		j := strings.Index(lower[start:], closeTag)
		if j < 0 {
			break
		}
		out = append(out, s[start:start+j])
		idx = start + j + len(closeTag)
	}
	return out
}
