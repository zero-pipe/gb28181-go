package server

import (
	"context"
	"fmt"
	"log"
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

	records *session.RecordWaiter
	presets *session.PresetWaiter
	invites *session.InviteManager
	controls *session.ControlWaiter
	status   *session.StatusWaiter
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
	}
	s.registerHandlers()
	if cfg.Port <= 0 || strings.TrimSpace(cfg.ID) == "" {
		log.Printf("[gb28181-go] SIP not fully configured (id/port)")
	} else if localIP == "" {
		log.Printf("[gb28181-go] WARNING: IP empty and auto-detect failed; INVITE may time out")
	} else {
		log.Printf("[gb28181-go] localIP=%s (Contact/Via)", localIP)
	}
	return s, nil
}

func (s *Server) SetHandlers(h Handlers) { s.handlers = h }

func (s *Server) SetLocalIP(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "0.0.0.0" || ip == "127.0.0.1" {
		return
	}
	if s.localIP == "" {
		s.localIP = ip
		log.Printf("[gb28181-go] localIP set: %s", ip)
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
	log.Printf("[gb28181-go] config applied id=%s domain=%s ip=%s port=%d",
		cfg.ID, cfg.Domain, s.localIP, cfg.Port)
}

func (s *Server) Config() Config   { return s.cfg }
func (s *Server) LocalIP() string  { return s.localIP }
func (s *Server) Domain() string   { return s.cfg.Domain }
func (s *Server) Records() *session.RecordWaiter   { return s.records }
func (s *Server) Presets() *session.PresetWaiter   { return s.presets }
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
}

// Start listens on UDP+TCP for cfg.Port. No-op if port <= 0.
func (s *Server) Start(ctx context.Context) error {
	if s.cfg.Port <= 0 {
		log.Printf("[gb28181-go] SIP not listening: port unset")
		return nil
	}
	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.Port)
	go func() {
		if err := s.srv.ListenAndServe(ctx, "udp", addr); err != nil && ctx.Err() == nil {
			log.Printf("[gb28181-go] sip udp error: %v", err)
		}
	}()
	go func() {
		if err := s.srv.ListenAndServe(ctx, "tcp", addr); err != nil && ctx.Err() == nil {
			log.Printf("[gb28181-go] sip tcp error: %v", err)
		}
	}()
	log.Printf("[gb28181-go] listening on %s (TCP+UDP)", addr)
	return nil
}

func (s *Server) handleBye(req *sip.Request, tx sip.ServerTransaction) {
	callID := ""
	if h := req.CallID(); h != nil {
		callID = h.Value()
	}
	removed := s.invites.RemoveByCallID(callID)
	log.Printf("[gb28181-go] BYE from=%s source=%s callID=%s removedSessions=%v",
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
