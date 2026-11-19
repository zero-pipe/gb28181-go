package server

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/digest"
)

const (
	// defaultRegisterExpires is the GB28181 default registration lifetime.
	defaultRegisterExpires = 86400
	// minRegisterExpires is the minimum lifetime a device may request (01-§1).
	minRegisterExpires = 3600
	// challengeTTL bounds how long a 401 challenge nonce stays acceptable.
	challengeTTL = 5 * time.Minute
	// beijingZone is the fixed UTC+8 zone used for the SIP Date header.
	// GB28181 SIP time sync (9.10) requires Beijing time regardless of host TZ.
	beijingZoneOffset = 8 * 3600
)

var beijingZone = time.FixedZone("CST", beijingZoneOffset)

func (s *Server) handleRegister(req *sip.Request, tx sip.ServerTransaction) {
	deviceID := extractSIPUser(req)
	if deviceID == "" {
		s.respond(tx, req, 400, "Bad Request")
		return
	}

	expires := s.parseExpires(req)
	ip, port := extractSourceAddr(req)

	password, known, ok := s.resolveCredentials(deviceID, tx, req)
	if !ok {
		return
	}
	if !s.authenticate(req, tx, deviceID, password) {
		return
	}

	if expires == 0 {
		s.handleUnregister(req, tx, deviceID)
		return
	}

	if s.redirect(req, tx, deviceID, expires) {
		return
	}

	ev := RegisterEvent{
		DeviceID:    deviceID,
		IP:          ip,
		Port:        port,
		Transport:   strings.ToUpper(string(req.Transport())),
		Expires:     expires,
		ServerID:    s.cfg.ServerID,
		IsNewDevice: !known,
	}
	if callID := req.CallID(); callID != nil {
		ev.CallID = callID.Value()
	}
	peer := Peer{DeviceID: deviceID, IP: ip, Port: port, LocalIP: "", SDPIP: ""}
	if lip := s.resolveInviteLocalIP(peer); lip != "" {
		ev.LocalIP = lip
	}

	if ev.Transport == "TCP" {
		s.tcpPeers.bind(deviceID, req.Source())
	}
	if s.handlers.Register != nil {
		if err := s.handlers.Register.OnRegister(context.Background(), ev); err != nil {
			s.logf("[gb28181-go] OnRegister %s: %v", deviceID, err)
			s.respond(tx, req, 500, "Server Error")
			return
		}
	}
	if s.handlers.Telemetry != nil {
		s.handlers.Telemetry.OnRegister(deviceID, time.Now())
	}

	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	// SIP time sync (9.10): Date carries Beijing time as yyyy-MM-ddTHH:mm:ss.SSS.
	res.AppendHeader(sip.NewHeader("Date", time.Now().In(beijingZone).Format("2006-01-02T15:04:05.000")))
	res.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(expires)))
	_ = tx.Respond(res)
	s.logf("[gb28181-go] registered: id=%s ip=%s:%d transport=%s expires=%d",
		deviceID, ip, port, ev.Transport, expires)
}

// redirect answers 302 Moved Temporarily when a redirect policy is configured
// (2022 §9.1.2.3). Returns true if it handled the request.
func (s *Server) redirect(req *sip.Request, tx sip.ServerTransaction, deviceID string, expires int) bool {
	if s.handlers.Redirect == nil {
		return false
	}
	contact, targetExpires, ok := s.handlers.Redirect.Redirect(deviceID)
	if !ok || strings.TrimSpace(contact) == "" {
		return false
	}
	if targetExpires <= 0 {
		targetExpires = expires
	}
	res := sip.NewResponseFromRequest(req, sip.StatusMovedTemporarily, "Moved Temporarily", nil)
	res.AppendHeader(sip.NewHeader("Contact", contact))
	res.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(targetExpires)))
	_ = tx.Respond(res)
	s.logf("[gb28181-go] register redirected: id=%s -> %s", deviceID, contact)
	return true
}

// resolveCredentials loads the device Digest password and whether it is known.
// On error it responds 500 and returns ok=false.
func (s *Server) resolveCredentials(deviceID string, tx sip.ServerTransaction, req *sip.Request) (string, bool, bool) {
	password := s.cfg.Password
	known := false
	if s.handlers.Auth != nil {
		p, k, err := s.handlers.Auth.ResolvePassword(deviceID)
		if err != nil {
			s.respond(tx, req, 500, "Server Error")
			return "", false, false
		}
		known = k
		if p != "" {
			password = p
		}
	}
	if !known && s.cfg.RequirePreRegister {
		s.logf("[gb28181-go] register rejected (not pre-registered): id=%s", deviceID)
		s.respond(tx, req, 403, "Forbidden")
		return "", false, false
	}
	if password == "" && !known {
		s.respond(tx, req, 403, "Forbidden")
		return "", false, false
	}
	return password, known, true
}

// authenticate runs the Digest challenge-response. It responds to the
// transaction itself (401 / 403) and returns false when auth does not pass.
// Used by both REGISTER and unregister so Expires:0 cannot bypass auth.
func (s *Server) authenticate(req *sip.Request, tx sip.ServerTransaction, deviceID, password string) bool {
	auth := req.GetHeader("Authorization")
	if auth == nil {
		s.challenge(tx, req, deviceID, false)
		return false
	}
	chVal, found := s.challenges.Load(deviceID)
	ch, _ := chVal.(digest.Challenge)
	if !found || ch.Expired(challengeTTL) {
		// No live challenge (e.g. server restarted) → re-challenge.
		s.challenge(tx, req, deviceID, true)
		return false
	}
	authParams := digest.ParseAuthorization(auth.Value())
	uri := authParams["uri"]
	if uri == "" {
		uri = req.Recipient.String()
	}
	if !digest.Verify(auth.Value(), "REGISTER", uri, deviceID, password, ch.Realm, ch.Nonce) {
		if authParams["nonce"] != ch.Nonce {
			// Stale/forged nonce: issue a fresh challenge instead of a hard 403.
			s.logf("[gb28181-go] register stale nonce: id=%s", deviceID)
			s.challenge(tx, req, deviceID, true)
			return false
		}
		s.logf("[gb28181-go] register auth failed: id=%s", deviceID)
		s.respond(tx, req, 403, "Forbidden")
		return false
	}
	s.challenges.Delete(deviceID)
	return true
}

func (s *Server) challenge(tx sip.ServerTransaction, req *sip.Request, deviceID string, stale bool) {
	ch := digest.NewChallenge(s.cfg.Domain)
	s.challenges.Store(deviceID, ch)
	value := ch.String()
	if stale {
		value = ch.StringStale()
	}
	res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
	res.AppendHeader(sip.NewHeader("WWW-Authenticate", value))
	_ = tx.Respond(res)
}

func (s *Server) handleUnregister(req *sip.Request, tx sip.ServerTransaction, deviceID string) {
	s.tcpPeers.unbind(deviceID)
	if s.handlers.Register != nil {
		_ = s.handlers.Register.OnUnregister(context.Background(), deviceID)
	}
	s.respond(tx, req, 200, "OK")
	s.logf("[gb28181-go] unregistered: id=%s", deviceID)
}

func extractSIPUser(req *sip.Request) string {
	if from := req.From(); from != nil {
		return from.Address.User
	}
	return ""
}

// parseExpires reads the requested registration lifetime, applying the GB28181
// default (86400) when absent and clamping to at least minRegisterExpires.
func (s *Server) parseExpires(req *sip.Request) int {
	v, ok := requestedExpires(req)
	if !ok {
		return s.cfg.ExpiresOrDefault()
	}
	if v == 0 {
		return 0 // explicit unregister
	}
	min := s.cfg.MinExpiresOrDefault()
	if v < min {
		return min
	}
	return v
}

func requestedExpires(req *sip.Request) (int, bool) {
	if h := req.GetHeader("Expires"); h != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(h.Value())); err == nil {
			return v, true
		}
	}
	if c := req.Contact(); c != nil && c.Address.UriParams != nil {
		if exp, ok := c.Address.UriParams.Get("expires"); ok && exp != "" {
			if v, err := strconv.Atoi(exp); err == nil {
				return v, true
			}
		}
	}
	return 0, false
}

func extractSourceAddr(req *sip.Request) (string, int) {
	host, portStr, err := net.SplitHostPort(req.Source())
	if err != nil {
		return req.Source(), 5060
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}
