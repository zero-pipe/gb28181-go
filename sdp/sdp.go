package sdp

import (
	"fmt"
	"strings"
	"time"
)

// DefaultMediaFormat is the f= line (附录-编码与SDP) for a PS video stream:
// video coding 2, resolution 5, 25 fps, 4 Mbps; audio G.711A, 8 kHz, mono.
// Devices use it to pick encoder parameters; unknown platform values are sent
// as the common defaults.
const DefaultMediaFormat = "v/2/5/25/1/4000a/1/8/1"

// BuildPlay builds a GB28181 real-time play INVITE SDP body.
func BuildPlay(deviceID, sdpIP string, port int, ssrc, streamMode, streamIdentification string) string {
	mediaProto, setup := mediaSetup(streamMode)
	streamAttr := ""
	if streamIdentification != "" {
		streamAttr = fmt.Sprintf("a=%s\r\n", streamIdentification)
	}
	return fmt.Sprintf(
		"v=0\r\no=%s 0 0 IN IP4 %s\r\ns=Play\r\nc=IN IP4 %s\r\nt=0 0\r\nm=video %d %s 96 98 97 99 100\r\na=recvonly\r\na=rtpmap:96 PS/90000\r\na=rtpmap:98 H264/90000\r\na=rtpmap:97 MPEG4/90000\r\na=rtpmap:99 H265/90000\r\na=rtpmap:100 H265/90000\r\n%s%sf=%s\r\ny=%s\r\n",
		deviceID, sdpIP, sdpIP, port, mediaProto, streamAttr, setup, DefaultMediaFormat, ssrc,
	)
}

// BuildPlayback builds a GB28181 playback INVITE SDP body.
func BuildPlayback(deviceID, channelID, sdpIP string, port int, ssrc, streamMode, startTime, endTime string) string {
	mediaProto, setup := mediaSetup(streamMode)
	tStart, tEnd := parseTimeRange(startTime, endTime)
	return fmt.Sprintf(
		"v=0\r\no=%s 0 0 IN IP4 %s\r\ns=Playback\r\nu=%s:0\r\nc=IN IP4 %s\r\nt=%d %d\r\nm=video %d %s 96 98 97 99 100\r\na=recvonly\r\na=rtpmap:96 PS/90000\r\na=rtpmap:98 H264/90000\r\na=rtpmap:97 MPEG4/90000\r\na=rtpmap:99 H265/90000\r\na=rtpmap:100 H265/90000\r\n%sf=%s\r\ny=%s\r\n",
		deviceID, sdpIP, channelID, sdpIP, tStart, tEnd, port, mediaProto, setup, DefaultMediaFormat, ssrc,
	)
}

// BuildDownload builds a GB28181 download INVITE SDP body.
func BuildDownload(deviceID, channelID, sdpIP string, port int, ssrc, streamMode, startTime, endTime string, downloadSpeed int) string {
	mediaProto, setup := mediaSetup(streamMode)
	if downloadSpeed <= 0 {
		downloadSpeed = 4
	}
	tStart, tEnd := parseTimeRange(startTime, endTime)
	return fmt.Sprintf(
		"v=0\r\no=%s 0 0 IN IP4 %s\r\ns=Download\r\nu=%s:0\r\nc=IN IP4 %s\r\nt=%d %d\r\nm=video %d %s 96 98 97 99 100\r\na=recvonly\r\na=rtpmap:96 PS/90000\r\na=rtpmap:98 H264/90000\r\na=rtpmap:97 MPEG4/90000\r\na=rtpmap:99 H265/90000\r\na=rtpmap:100 H265/90000\r\n%sf=%s\r\na=downloadspeed:%d\r\ny=%s\r\n",
		deviceID, sdpIP, channelID, sdpIP, tStart, tEnd, port, mediaProto, setup, DefaultMediaFormat, downloadSpeed, ssrc,
	)
}

// ParseAnswerMedia extracts media host and port from an INVITE 200 OK SDP
// (used for TCP-ACTIVE connect-back).
func ParseAnswerMedia(sdpBody string) (host string, port int, err error) {
	sdpBody = strings.ReplaceAll(sdpBody, "\r\n", "\n")
	for _, line := range strings.Split(sdpBody, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "c=IN IP4 ") {
			host = strings.TrimSpace(strings.TrimPrefix(line, "c=IN IP4 "))
		}
		if strings.HasPrefix(line, "m=video ") {
			var mediaPort int
			if _, scanErr := fmt.Sscanf(line, "m=video %d", &mediaPort); scanErr == nil && mediaPort > 0 {
				port = mediaPort
			}
		}
	}
	if host == "" || port <= 0 {
		return "", 0, fmt.Errorf("answer SDP missing c= or m=video port")
	}
	return host, port, nil
}

// ExtractVideoPort returns the m=video port from an SDP body, or 0.
func ExtractVideoPort(sdpBody string) int {
	for _, line := range strings.Split(sdpBody, "\n") {
		if strings.HasPrefix(line, "m=video ") {
			var port int
			if _, err := fmt.Sscanf(line, "m=video %d", &port); err == nil {
				return port
			}
		}
	}
	return 0
}

func mediaSetup(streamMode string) (mediaProto, setup string) {
	mediaProto = "RTP/AVP"
	switch streamMode {
	case "TCP-ACTIVE":
		mediaProto = "TCP/RTP/AVP"
		setup = "a=setup:active\r\na=connection:new\r\n"
	case "TCP-PASSIVE":
		mediaProto = "TCP/RTP/AVP"
		setup = "a=setup:passive\r\na=connection:new\r\n"
	}
	return mediaProto, setup
}

func parseTimeRange(start, end string) (int64, int64) {
	const layout = "2006-01-02 15:04:05"
	ts, _ := time.ParseInLocation(layout, start, time.Local)
	te, _ := time.ParseInLocation(layout, end, time.Local)
	return ts.Unix(), te.Unix()
}
