package server

import (
	"context"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

// AuthResolver supplies Digest credentials and pre-register policy.
type AuthResolver interface {
	// ResolvePassword returns the Digest password for deviceID.
	// known=false means the device is not in the host registry.
	ResolvePassword(deviceID string) (password string, known bool, err error)
}

// RegisterHandler handles REGISTER success / unregister.
type RegisterHandler interface {
	OnRegister(ctx context.Context, ev RegisterEvent) error
	OnUnregister(ctx context.Context, deviceID string) error
}

// MessageHandler handles inbound MANSCDP Notify/Response (except RecordInfo/PresetQuery waiters).
type MessageHandler interface {
	// DeviceKnown reports whether MESSAGE from deviceID should be accepted (else 404).
	DeviceKnown(deviceID string) bool
	OnKeepalive(ctx context.Context, deviceID, ip string, port int) error
	OnCatalog(ctx context.Context, deviceID string, items []manscdp.CatalogItem) error
	OnDeviceInfo(ctx context.Context, deviceID, name, manufacturer, model, firmware string) error
	OnDeviceStatus(ctx context.Context, deviceID string, status *manscdp.DeviceStatus) error
	OnMediaStatus(ctx context.Context, deviceID string, status *manscdp.MediaStatusNotify) error
	OnDeviceControlResult(ctx context.Context, deviceID, sn, result string) error
	OnAlarm(ctx context.Context, deviceID, channelID string, alarm *manscdp.AlarmNotify) error
	OnMobilePosition(ctx context.Context, deviceID, channelID string, pos *manscdp.MobilePositionNotify) error
}

// TelemetryHook is optional side-effect hooks (e.g. Redis timestamps).
type TelemetryHook interface {
	OnRegister(deviceID string, ts time.Time)
	OnKeepalive(deviceID string, ts time.Time)
}

// CascadeInboundHandler handles inbound control/media from an upstream platform.
type CascadeInboundHandler interface {
	UpstreamKnown(upstreamGBID string) bool
	OnDeviceControl(ctx context.Context, ev InboundControlEvent) error
	OnInvite(ctx context.Context, ev InboundInviteEvent) (answerSDP []byte, err error)
	OnInviteEnd(ctx context.Context, callID string) error
}

// InboundControlEvent is a superior MESSAGE DeviceControl.
type InboundControlEvent struct {
	UpstreamGBID string
	ChannelGBID  string
	SN           string
	PTZCmd       string
	Raw          []byte
}

// InboundInviteEvent is a superior INVITE for a shared catalog channel.
type InboundInviteEvent struct {
	UpstreamGBID string
	ChannelGBID  string
	CallID       string
	Subject      string
	OfferSDP     []byte
	SourceIP     string
	SourcePort   int
}

// Handlers bundles callbacks for the SIP server.
type Handlers struct {
	Auth           AuthResolver
	Register       RegisterHandler
	Message        MessageHandler
	Telemetry      TelemetryHook
	CascadeInbound CascadeInboundHandler
}

// RegisterEvent carries REGISTER-derived fields for the host to persist.
type RegisterEvent struct {
	DeviceID       string
	IP             string
	Port           int
	Transport      string
	Expires        int
	CallID         string
	LocalIP        string
	ServerID       string
	IsNewDevice    bool // known=false before register (auto-create candidate)
}
