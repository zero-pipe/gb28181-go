package server

import (
	"context"
	"log"

	"github.com/emiago/sipgo/sip"
)

func (s *Server) handleInvite(req *sip.Request, tx sip.ServerTransaction) {
	if s.handlers.CascadeInbound == nil {
		s.respond(tx, req, 405, "Method Not Allowed")
		return
	}
	upstreamGBID := extractSIPUser(req)
	if !s.handlers.CascadeInbound.UpstreamKnown(upstreamGBID) {
		s.respond(tx, req, 404, "Not Found")
		return
	}
	channelGBID := req.Recipient.User
	if channelGBID == "" {
		s.respond(tx, req, 400, "Bad Request")
		return
	}
	callID := ""
	if h := req.CallID(); h != nil {
		callID = h.Value()
	}
	subject := ""
	if h := req.GetHeader("Subject"); h != nil {
		subject = h.Value()
	}
	srcIP, srcPort := extractSourceAddr(req)
	answer, err := s.handlers.CascadeInbound.OnInvite(context.Background(), InboundInviteEvent{
		UpstreamGBID: upstreamGBID,
		ChannelGBID:  channelGBID,
		CallID:       callID,
		Subject:      subject,
		OfferSDP:     append([]byte(nil), req.Body()...),
		SourceIP:     srcIP,
		SourcePort:   srcPort,
	})
	if err != nil {
		log.Printf("[gb28181-go] inbound INVITE failed upstream=%s channel=%s err=%v", upstreamGBID, channelGBID, err)
		s.respond(tx, req, 404, "Not Found")
		return
	}
	res := sip.NewResponseFromRequest(req, 200, "OK", answer)
	res.AppendHeader(sip.NewHeader("Content-Type", "APPLICATION/SDP"))
	if err := tx.Respond(res); err != nil {
		log.Printf("[gb28181-go] inbound INVITE respond err=%v", err)
	}
}
