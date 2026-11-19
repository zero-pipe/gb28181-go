package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/session"
)

// Server is a GB28181 platform-side SIP UA (REGISTER/MESSAGE/INVITE).
type Server struct {
	cfg      Config
	localIP  string
	ua       *sipgo.UserAgent
	srv      *sipgo.Server
	client   *sipgo.Client
	handlers Handlers

	challenges sync.Map
	sn         atomic.Int64
	infoCSeq   atomic.Int64

	records  *session.RecordWaiter
	presets  *session.PresetWaiter
	invites  *session.InviteManager
	controls *session.ControlWaiter
	status   *session.StatusWaiter
	queries  *session.QueryWaiter
	tcpPeers *tcpPeerRegistry
}

// New creates a SIP server. handlers.Auth and handlers.Register/Message should be set
// before Start for full platform behavior.
func New(cfg Config, handlers Handlers) (*Server, error) {
	localIP := strings.TrimSpace(cfg.IP)
	if localIP == "" || localIP == "0.0.0.0" || localIP == "127.0.0.1" {
		localIP = GuessLocalIP()
	}
	uaName := strings.TrimSpace(cfg.UserAgent)
	if uaName == "" {
		uaName = strings.TrimSpace(cfg.ID)
	}
	if uaName == "" {
		uaName = "gb28181-go"
	}
	uaOpts := []sipgo.UserAgentOption{sipgo.WithUserAgent(uaName)}
	if localIP != "" {
		uaOpts = append(uaOpts, sipgo.WithUserAgentHostname(localIP))
	}
	ua, err := sipgo.NewUA(uaOpts...)
	if err != nil {
		return nil, err
	}
	srv, err := sipgo.NewServer(ua)
	if err != nil {
		return nil, err
	}
	var clientOpts []sipgo.ClientOption
	if localIP != "" && cfg.Port > 0 {
		clientOpts = append(clientOpts,
			sipgo.WithClientHostname(localIP),
			sipgo.WithClientPort(cfg.Port),
			sipgo.WithClientNAT(),
		)
	}
	client, err := sipgo.NewClient(ua, clientOpts...)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:      cfg,
		localIP:  localIP,
		ua:       ua,
		srv:      srv,
		client:   client,
		handlers: handlers,
		records:  session.NewRecordWaiter(),
		presets:  session.NewPresetWaiter(),
		invites:  session.NewInviteManager(),
		controls: session.NewControlWaiter(),
		status:   session.NewStatusWaiter(),
		queries:  session.NewQueryWaiter(),
	}
	s.registerHandlers()
	if cfg.Port <= 0 || strings.TrimSpace(cfg.ID) == "" {
		s.logf("[gb28181-go] SIP not fully configured (id/port)")
	} else if localIP == "" {
		s.logf("[gb28181-go] WARNING: IP empty and auto-detect failed; INVITE may time out")
	} else {
		s.logf("[gb28181-go] localIP=%s (Contact/Via)", localIP)
	}
	return s, nil
}

// logf writes library diagnostics through the configured Logger.
func (s *Server) logf(format string, args ...any) {
	logger := s.cfg.Logger
	if logger == nil {
		logger = log.Default()
	}
	logger.Printf(format, args...)
}

func (s *Server) SetHandlers(h Handlers) { s.handlers = h }

func (s *Server) SetLocalIP(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "0.0.0.0" || ip == "127.0.0.1" {
		return
	}
	if s.localIP == "" {
		s.localIP = ip
		s.logf("[gb28181-go] localIP set: %s", ip)
	}
}

// ApplyConfig hot-updates identity/password/domain/reachable IP.
// Listen port changes require process restart.
func (s *Server) ApplyConfig(cfg Config) {
	s.cfg = cfg
	ip := strings.TrimSpace(cfg.IP)
	if ip != "" && ip != "0.0.0.0" && ip != "127.0.0.1" {
		s.localIP = ip
	}
	s.logf("[gb28181-go] config applied id=%s domain=%s ip=%s port=%d",
		cfg.ID, cfg.Domain, s.localIP, cfg.Port)
}

func (s *Server) Config() Config                   { return s.cfg }
func (s *Server) LocalIP() string                  { return s.localIP }
func (s *Server) Domain() string                   { return s.cfg.Domain }
func (s *Server) Records() *session.RecordWaiter   { return s.records }
func (s *Server) Presets() *session.PresetWaiter   { return s.presets }
func (s *Server) Queries() *session.QueryWaiter    { return s.queries }
func (s *Server) Invites() *session.InviteManager  { return s.invites }
func (s *Server) Controls() *session.ControlWaiter { return s.controls }
func (s *Server) Status() *session.StatusWaiter    { return s.status }

func (s *Server) NextSN() string {
	return fmt.Sprintf("%d", s.sn.Add(1))
}

func (s *Server) NextInfoCSeq() int {
	return int(s.infoCSeq.Add(1))
}

func (s *Server) registerHandlers() {
	s.srv.OnRegister(s.handleRegister)
	s.srv.OnMessage(s.handleMessage)
	s.srv.OnNotify(s.handleMessage) // catalog push / subscription notify (抓包 11)
	s.srv.OnInvite(s.handleInvite)
	s.srv.OnBye(s.handleBye)
	s.srv.OnOptions(s.handleOptions)
	s.srv.OnSubscribe(s.handleSubscribe)
	s.srv.OnInfo(s.handleInfo)
}

// handleInfo forwards superior playback control (MANSRTSP over SIP INFO).
func (s *Server) handleInfo(req *sip.Request, tx sip.ServerTransaction) {
	callID := ""
	if c := req.CallID(); c != nil {
		callID = c.Value()
	}
	ct := ""
	if h := req.GetHeader("Content-Type"); h != nil {
		ct = strings.TrimSpace(h.Value())
	}
	ev := InboundInfoEvent{
		UpstreamGBID: extractSIPUser(req),
		CallID:       callID,
		ChannelGBID:  req.Recipient.User,
		ContentType:  ct,
		Body:         append([]byte(nil), req.Body()...),
	}
	if s.handlers.CascadeInbound != nil {
		if err := s.handlers.CascadeInbound.OnInfo(context.Background(), ev); err != nil {
			s.logf("[gb28181-go] INFO forward failed: callID=%s err=%v", callID, err)
		}
	}
	s.respond(tx, req, 200, "OK")
}

// handleSubscribe accepts an upstream platform's SUBSCRIBE (Catalog/Alarm/
// MobilePosition) when the platform is known, replying with the granted Expires.
func (s *Server) handleSubscribe(req *sip.Request, tx sip.ServerTransaction) {
	deviceID := extractSIPUser(req)
	if deviceID == "" {
		s.respond(tx, req, 400, "Bad Request")
		return
	}
	event := ""
	if h := req.GetHeader("Event"); h != nil {
		event = strings.TrimSpace(h.Value())
	}
	expires := 3600
	if h := req.GetHeader("Expires"); h != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(h.Value())); err == nil {
			expires = v
		}
	}
	callID := ""
	if c := req.CallID(); c != nil {
		callID = c.Value()
	}
	ip, port := extractSourceAddr(req)
	ev := InboundSubscribeEvent{
		UpstreamGBID: deviceID, Event: event, Expires: expires,
		CallID: callID, SourceIP: ip, SourcePort: port,
	}
	if s.handlers.CascadeInbound == nil || !s.handlers.CascadeInbound.UpstreamKnown(deviceID) {
		s.logf("[gb28181-go] subscribe rejected (unknown upstream): %s event=%s", deviceID, event)
		s.respond(tx, req, 403, "Forbidden")
		return
	}
	granted, accepted, err := s.handlers.CascadeInbound.OnSubscribe(context.Background(), ev)
	if err != nil || !accepted {
		s.logf("[gb28181-go] subscribe rejected: upstream=%s event=%s err=%v", deviceID, event, err)
		s.respond(tx, req, 403, "Forbidden")
		return
	}
	if granted <= 0 {
		granted = expires
	}
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	if event != "" {
		res.AppendHeader(sip.NewHeader("Event", event))
	}
	res.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(granted)))
	_ = tx.Respond(res)
	s.logf("[gb28181-go] subscribe accepted: upstream=%s event=%s expires=%d", deviceID, event, granted)
}

// allowedMethods is advertised in the Allow header of an OPTIONS response.
const allowedMethods = "REGISTER, MESSAGE, INVITE, ACK, CANCEL, BYE, OPTIONS, INFO, SUBSCRIBE, NOTIFY"

// handleOptions answers a capability probe (9.2.3 / RFC 3261 §11).
func (s *Server) handleOptions(req *sip.Request, tx sip.ServerTransaction) {
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	res.AppendHeader(sip.NewHeader("Allow", allowedMethods))
	res.AppendHeader(sip.NewHeader("Accept", "application/sdp, application/MANSCDP+xml"))
	_ = tx.Respond(res)
}

// Start listens on UDP+TCP for cfg.Port. No-op if port <= 0.
func (s *Server) Start(ctx context.Context) error {
	if s.cfg.Port <= 0 {
		s.logf("[gb28181-go] SIP not listening: port unset")
		return nil
	}
	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.Port)
	go func() {
		if err := s.srv.ListenAndServe(ctx, "udp", addr); err != nil && ctx.Err() == nil {
			s.logf("[gb28181-go] sip udp error: %v", err)
		}
	}()
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			if ctx.Err() == nil {
				s.logf("[gb28181-go] sip tcp listen error: %v", err)
			}
			return
		}
		go func() {
			<-ctx.Done()
			_ = ln.Close()
		}()
		listener := &trackingListener{Listener: ln, notify: s.onTCPClosed}
		if err := s.srv.ServeTCP(listener); err != nil && ctx.Err() == nil {
			s.logf("[gb28181-go] sip tcp error: %v", err)
		}
	}()
	s.logf("[gb28181-go] listening on %s (TCP+UDP)", addr)
	return nil
}

// onTCPClosed is the registry callback used by the tracking listener.
func (s *Server) onTCPClosed(remote string) {
	if deviceID := s.tcpPeers.closed(remote); deviceID != "" {
		s.logf("[gb28181-go] tcp closed: device=%s addr=%s", deviceID, remote)
		if s.handlers.Transport != nil {
			_ = s.handlers.Transport.OnTransportClosed(context.Background(), deviceID)
		}
	}
}

func (s *Server) handleBye(req *sip.Request, tx sip.ServerTransaction) {
	callID := ""
	if h := req.CallID(); h != nil {
		callID = h.Value()
	}
	removed := s.invites.RemoveByCallID(callID)
	s.logf("[gb28181-go] BYE from=%s source=%s callID=%s removedSessions=%v",
		extractSIPUser(req), req.Source(), callID, removed)
	if s.handlers.CascadeInbound != nil && callID != "" {
		_ = s.handlers.CascadeInbound.OnInviteEnd(context.Background(), callID)
	}
	s.respond(tx, req, 200, "OK")
}

func (s *Server) respond(tx sip.ServerTransaction, req *sip.Request, code int, reason string) {
	res := sip.NewResponseFromRequest(req, code, reason, nil)
	_ = tx.Respond(res)
}
