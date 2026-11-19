package sdp

import (
	"strconv"
	"strings"
	"time"
)

// OfferKind is the s= session type of an inbound offer.
const (
	OfferPlay     = "Play"
	OfferPlayback = "Playback"
	OfferDownload = "Download"
)

// OfferInfo holds the fields a superior's INVITE offer carries that affect how
// we serve the stream (9.2 / 9.8 / 9.9).
type OfferInfo struct {
	Kind      string // Play / Playback / Download
	StartTime string // "2006-01-02 15:04:05" for Playback/Download
	EndTime   string
	SSRC      string
	Setup     string // actpass/active/passive (TCP)
}

const timeLayout = "2006-01-02 15:04:05"

// ParseOffer extracts session type, time range and SSRC from an offer SDP.
func ParseOffer(body []byte) OfferInfo {
	info := OfferInfo{Kind: OfferPlay}
	normalized := strings.ReplaceAll(string(body), "\r\n", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "s="):
			v := strings.TrimSpace(strings.TrimPrefix(line, "s="))
			switch strings.ToLower(v) {
			case "playback":
				info.Kind = OfferPlayback
			case "download":
				info.Kind = OfferDownload
			case "play":
				info.Kind = OfferPlay
			}
		case strings.HasPrefix(line, "t="):
			fields := strings.Fields(strings.TrimPrefix(line, "t="))
			if len(fields) == 2 {
				if ts, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
					info.StartTime = time.Unix(ts, 0).Format(timeLayout)
				}
				if te, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					info.EndTime = time.Unix(te, 0).Format(timeLayout)
				}
			}
		case strings.HasPrefix(line, "y="):
			info.SSRC = strings.TrimSpace(strings.TrimPrefix(line, "y="))
		case strings.HasPrefix(line, "a=setup:"):
			info.Setup = strings.TrimSpace(strings.TrimPrefix(line, "a=setup:"))
		}
	}
	return info
}
