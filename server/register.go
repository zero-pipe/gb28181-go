package server

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/digest"
)

func (s *Server) handleRegister(req *sip.Request, tx sip.ServerTransaction) {
	deviceID := extractSIPUser(req)
	if deviceID == "" {
		s.respond(tx, req, 400, "Bad Request")
		return
	}

	expires := parseExpires(req)
	ip, port := extractSourceAddr(req)

	if expires == 0 {
		s.handleUnregister(req, tx, deviceID)
		return
	}

	password := s.cfg.Password
	known := false
	if s.handlers.Auth != nil {
		p, k, err := s.handlers.Auth.ResolvePassword(deviceID)
		if err != nil {
			s.respond(tx, req, 500, "Server Error")
			return
		}
		known = k
		if p != "" {
			password = p
		}
	}
	if !known && s.cfg.RequirePreRegister {
		log.Printf("[gb28181-go] register rejected (not pre-registered): id=%s ip=%s:%d", deviceID, ip, port)
		s.respond(tx, req, 403, "Forbidden")
		return
	}
	if password == "" && !known {
		s.respond(tx, req, 403, "Forbidden")
		return
	}

	auth := req.GetHeader("Authorization")
	if auth == nil {
		ch := digest.NewChallenge(s.cfg.Domain)
		s.challenges.Store(deviceID, ch)
		res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
		res.AppendHeader(sip.NewHeader("WWW-Authenticate", ch.String()))
		_ = tx.Respond(res)
		return
	}

	chVal, _ := s.challenges.Load(deviceID)
	ch, _ := chVal.(digest.Challenge)
	authParams := digest.ParseAuthorization(auth.Value())
	uri := authParams["uri"]
	if uri == "" {
		uri = req.Recipient.String()
	}
	if !digest.Verify(auth.Value(), "REGISTER", uri, deviceID, password, ch.Realm, ch.Nonce) {
		log.Printf("[gb28181-go] register auth failed: id=%s ip=%s:%d", deviceID, ip, port)
		s.respond(tx, req, 403, "Forbidden")
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

	if s.handlers.Register != nil {
		if err := s.handlers.Register.OnRegister(context.Background(), ev); err != nil {
			log.Printf("[gb28181-go] OnRegister %s: %v", deviceID, err)
			s.respond(tx, req, 500, "Server Error")
			return
		}
	}
	if s.handlers.Telemetry != nil {
		s.handlers.Telemetry.OnRegister(deviceID, time.Now())
	}

	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	res.AppendHeader(sip.NewHeader("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")))
	res.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(expires)))
	_ = tx.Respond(res)
	log.Printf("[gb28181-go] registered: id=%s ip=%s:%d transport=%s expires=%d",
		deviceID, ip, port, ev.Transport, expires)
}

func (s *Server) handleUnregister(req *sip.Request, tx sip.ServerTransaction, deviceID string) {
	if s.handlers.Register != nil {
		_ = s.handlers.Register.OnUnregister(context.Background(), deviceID)
	}
	s.respond(tx, req, 200, "OK")
}

func extractSIPUser(req *sip.Request) string {
	if from := req.From(); from != nil {
		return from.Address.User
	}
	return ""
}

func parseExpires(req *sip.Request) int {
	if h := req.GetHeader("Expires"); h != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(h.Value())); err == nil {
			return v
		}
	}
	if c := req.Contact(); c != nil && c.Address.UriParams != nil {
		if exp, ok := c.Address.UriParams.Get("expires"); ok && exp != "" {
			if v, err := strconv.Atoi(exp); err == nil {
				return v
			}
		}
	}
	return 3600
}

func extractSourceAddr(req *sip.Request) (string, int) {
	host, portStr, err := net.SplitHostPort(req.Source())
	if err != nil {
		return req.Source(), 5060
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}
