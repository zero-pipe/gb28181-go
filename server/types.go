package server

// Config is the GB28181 platform SIP identity and listen settings.
type Config struct {
	// ID is the platform 20-digit GB ID (SIP From user).
	ID string
	// Domain is the SIP realm / domain (often first 10 digits of ID).
	Domain string
	// Password is the default Digest password for devices.
	Password string
	// IP is the platform reachable IP used in Contact/Via (empty = auto-detect).
	IP string
	// Port is the SIP listen port (UDP+TCP).
	Port int
	// RequirePreRegister rejects REGISTER from unknown device IDs.
	RequirePreRegister bool
	// UserAgent is the SIP User-Agent product name (optional).
	UserAgent string
	// ServerID is an opaque host identifier echoed into RegisterEvent (optional).
	ServerID string
}

// Peer is a SIP endpoint (device or channel target) without host-app domain types.
type Peer struct {
	DeviceID       string
	IP             string
	Port           int
	Password       string
	LocalIP        string
	SDPIP          string
	Transport      string
	HostAddress    string
	Expires        int
	RegisterCallID string
}

// InviteTarget identifies the media channel for INVITE (usually channel GB ID).
type InviteTarget struct {
	ChannelID string // GB channel / device ID in Request-URI user part
}
