package ssrc

import "fmt"

// Play returns a GB28181 play SSRC: 0 + domain[3:8] + 4-digit seq.
func Play(domain string, seq int) string {
	part := domain
	if len(domain) >= 8 {
		part = domain[3:8]
	}
	return fmt.Sprintf("0%s%04d", part, seq%10000)
}

// Playback returns a GB28181 playback/download SSRC: 1 + domain[3:8] + 4-digit
// seq. Historical media uses a leading 1 so devices and superiors can tell it
// apart from real-time (附录-编码与SDP).
func Playback(domain string, seq int) string {
	part := domain
	if len(domain) >= 8 {
		part = domain[3:8]
	}
	return fmt.Sprintf("1%s%04d", part, seq%10000)
}
