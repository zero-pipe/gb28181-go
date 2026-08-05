package transport

import "strings"

// Stream modes for INVITE SDP media transport.
// Independent of SIP signaling transport (UDP/TCP on port 5060).
const (
	UDP        = "UDP"
	TCPActive  = "TCP-ACTIVE"
	TCPPassive = "TCP-PASSIVE"
)

// Normalize returns the RTP media transport for INVITE SDP.
// Empty or unknown values default to UDP.
func Normalize(mode string) string {
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case TCPActive:
		return TCPActive
	case TCPPassive:
		return TCPPassive
	case UDP:
		return UDP
	default:
		return UDP
	}
}
