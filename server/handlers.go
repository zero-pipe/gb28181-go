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
	// OnCatalog receives a Catalog list from a device. incremental is true for a
	// NOTIFY (append/update/delete per item.Event) and false for a full query
	// response (replace the whole set).
	OnCatalog(ctx context.Context, deviceID string, items []manscdp.CatalogItem, incremental bool) error
	OnDeviceInfo(ctx context.Context, deviceID, name, manufacturer, model, firmware string) error
	OnDeviceStatus(ctx context.Context, deviceID string, status *manscdp.DeviceStatus) error
	OnMediaStatus(ctx context.Context, deviceID string, status *manscdp.MediaStatusNotify) error
	OnDeviceControlResult(ctx context.Context, deviceID, sn, result string) error
	OnAlarm(ctx context.Context, deviceID, channelID string, alarm *manscdp.AlarmNotify) error
	OnMobilePosition(ctx context.Context, deviceID, channelID string, pos *manscdp.MobilePositionNotify) error
	OnPTZPosition(ctx context.Context, deviceID, channelID string, pos *manscdp.PTZPosition) error
	OnSnapShotFinished(ctx context.Context, deviceID string, ev *manscdp.SnapShotFinished) error
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
	// OnSubscribe handles an inbound SUBSCRIBE from an upstream platform
	// (Catalog/Alarm/MobilePosition). Return the granted Expires in seconds;
	// accepted=false answers 403.
	OnSubscribe(ctx context.Context, ev InboundSubscribeEvent) (expires int, accepted bool, err error)
	// OnQuery handles an inbound MANSCDP Query from an upstream platform.
	// Return ok=false to answer an empty 200.
	OnQuery(ctx context.Context, ev InboundQueryEvent) (body []byte, ok bool)
	// OnInfo forwards a superior INFO (MANSRTSP playback control) to the leaf
	// session associated with the dialog.
	OnInfo(ctx context.Context, ev InboundInfoEvent) error
}

// InboundInfoEvent is a superior INFO message (playback control).
type InboundInfoEvent struct {
	UpstreamGBID string
	CallID       string
	ChannelGBID  string
	ContentType  string
	Body         []byte
}

// InboundSubscribeEvent is a superior SUBSCRIBE.
type InboundSubscribeEvent struct {
	UpstreamGBID  string
	Event         string // Catalog / Alarm / MobilePosition
	Expires       int
	CallID        string
	SourceIP      string
	SourcePort    int
}

// InboundQueryEvent is a superior MANSCDP Query (e.g. Catalog).
type InboundQueryEvent struct {
	UpstreamGBID string
	CmdType      string // Catalog / DeviceInfo / DeviceStatus / ...
	SN           string
	DeviceID     string
	Raw          []byte
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

// RedirectPolicy optionally redirects an authenticated REGISTER to another
// server (2022 §9.1.2.3). Returning ok=false serves the registration locally.
// contact must be a SIP URI: sip:<server-id>@<ip>:<port>.
type RedirectPolicy interface {
	Redirect(deviceID string) (contact string, expires int, ok bool)
}

// TransportHandler is notified when a device's TCP registration connection closes.
type TransportHandler interface {
	OnTransportClosed(ctx context.Context, deviceID string) error
}

// Handlers bundles callbacks for the SIP server.
type Handlers struct {
	Auth           AuthResolver
	Register       RegisterHandler
	Message        MessageHandler
	Telemetry      TelemetryHook
	CascadeInbound CascadeInboundHandler
	Redirect       RedirectPolicy
	Transport      TransportHandler
}

// RegisterEvent carries REGISTER-derived fields for the host to persist.
type RegisterEvent struct {
	DeviceID    string
	IP          string
	Port        int
	Transport   string
	Expires     int
	CallID      string
	LocalIP     string
	ServerID    string
	IsNewDevice bool // known=false before register (auto-create candidate)
}
