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
