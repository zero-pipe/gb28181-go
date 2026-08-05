package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/sdp"
	"github.com/zero-pipe/gb28181-go/session"
)

// SendInvitePlay starts a real-time play INVITE dialog.
// tcpConnect is used for TCP-ACTIVE after 200 OK; onOK receives the 200 response optionally.
func (s *Server) SendInvitePlay(
	peer Peer,
	target InviteTarget,
	sdpBody, ssrc, stream, streamMode string,
	tcpConnect func(host string, port int) error,
	onOK func(*sip.Response),
) error {
	dialog, err := s.inviteDialog(peer, target.ChannelID, sdpBody, ssrc)
	if err != nil {
		return err
	}
	if strings.EqualFold(streamMode, "TCP-ACTIVE") && tcpConnect != nil && dialog.InviteResponse != nil {
		host, port, parseErr := sdp.ParseAnswerMedia(string(dialog.InviteResponse.Body()))
		if parseErr != nil {
			log.Printf("[gb28181-go] TCP-ACTIVE parse 200 OK SDP failed: %v", parseErr)
			dialog.Close()
			return parseErr
		}
		log.Printf("[gb28181-go] TCP-ACTIVE connect -> %s:%d", host, port)
		if err := tcpConnect(host, port); err != nil {
			log.Printf("[gb28181-go] TCP-ACTIVE connect failed: %v", err)
			dialog.Close()
			return err
		}
	}
	s.invites.Put(stream, &session.InviteSession{
		DeviceID: peer.DeviceID, ChannelID: target.ChannelID,
		IP: peer.IP, Port: peer.Port,
		Stream: stream, App: "live", Type: session.SessionPlay,
		Dialog: dialog, StartedAt: time.Now(),
	})
	log.Printf("[gb28181-go] invite session stored stream=%s", stream)
	if onOK != nil && dialog.InviteResponse != nil {
		onOK(dialog.InviteResponse)
	}
	return nil
}

// SendInviteSession starts playback/download INVITE.
func (s *Server) SendInviteSession(
	peer Peer,
	target InviteTarget,
	sdpBody, ssrc, stream string,
	sessionType session.SessionType,
	startTime, endTime string,
	downloadSpeed int,
) error {
	dialog, err := s.inviteDialog(peer, target.ChannelID, sdpBody, ssrc)
	if err != nil {
		return err
	}
	s.invites.Put(stream, &session.InviteSession{
		DeviceID: peer.DeviceID, ChannelID: target.ChannelID,
		IP: peer.IP, Port: peer.Port,
		Stream: stream, App: "live", Type: sessionType,
		Dialog: dialog, StartTime: startTime, EndTime: endTime,
		DownloadSpeed: downloadSpeed, StartedAt: time.Now(),
	})
	return nil
}

// SendPlaybackControl sends MANSRTSP INFO on an active dialog.
func (s *Server) SendPlaybackControl(stream, content string) error {
	sess, ok := s.invites.Get(stream)
	if !ok {
		return session.ErrSessionNotFound
	}
	recipient := sess.Dialog.InviteRequest.Recipient
	req := sip.NewRequest(sip.INFO, recipient)
	req.SetBody([]byte(content))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSRTSP"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := sess.Dialog.Do(ctx, req)
	return err
}

// CloseInviteSession sends BYE and removes the session.
func (s *Server) CloseInviteSession(stream string) error {
	sess, ok := s.invites.Get(stream)
	if !ok {
		return nil
	}
	defer s.invites.Remove(stream)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sess.Dialog.Bye(ctx); err != nil {
		_ = sess.Dialog.Close()
	}
	return nil
}

func (s *Server) inviteDialog(peer Peer, channelID, sdpBody, ssrc string) (*sipgo.DialogClientSession, error) {
	localIP := s.resolveInviteLocalIP(peer)
	if localIP == "" {
		return nil, fmt.Errorf("platform IP unset: INVITE Contact needs a reachable IP")
	}
	dialogUA := &sipgo.DialogUA{
		Client: s.client,
		ContactHDR: sip.ContactHeader{
			Address: sip.Uri{User: s.cfg.ID, Host: localIP, Port: s.cfg.Port},
		},
		// EasyGBD etc. often put Contact without port (defaults to 5060). On same host that
		// routes ACK back to the platform SIP port → 405, device never gets ACK / never pushes.
		RewriteContact: true,
	}
	recipient := sip.Uri{User: channelID, Host: peer.IP, Port: peer.Port}
	subject := fmt.Sprintf("%s:%s,%s:0", channelID, ssrc, s.cfg.ID)
	log.Printf("[gb28181-go] INVITE %s@%s:%d Subject=%s localIP=%s sdpPort=%d",
		channelID, peer.IP, peer.Port, subject, localIP, sdp.ExtractVideoPort(sdpBody))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sess, err := dialogUA.Invite(ctx, recipient, []byte(sdpBody),
		sip.NewHeader("Content-Type", "APPLICATION/SDP"),
		sip.NewHeader("Subject", subject),
	)
	if err != nil {
		log.Printf("[gb28181-go] INVITE send error: %v", err)
		return nil, err
	}
	if err := sess.WaitAnswer(ctx, sipgo.AnswerOptions{}); err != nil {
		log.Printf("[gb28181-go] INVITE no 200 OK: %v", err)
		sess.Close()
		return nil, err
	}
	log.Printf("[gb28181-go] INVITE 200 OK status=%d source=%s", sess.InviteResponse.StatusCode, sess.InviteResponse.Source())
	if err := sess.Ack(ctx); err != nil {
		log.Printf("[gb28181-go] ACK error: %v", err)
		sess.Close()
		return nil, err
	}
	log.Printf("[gb28181-go] ACK sent -> %s (RewriteContact)", sess.InviteResponse.Source())
	return sess, nil
}
