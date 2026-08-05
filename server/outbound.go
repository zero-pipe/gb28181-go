package server

import (
	"strconv"

	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/manscdp"
	"github.com/zero-pipe/gb28181-go/ptz"
	"github.com/zero-pipe/gb28181-go/session"
)

func (s *Server) sendMessage(peer Peer, body string) error {
	if err := s.requirePeerAddr(peer); err != nil {
		return err
	}
	recipient := sip.Uri{User: peer.DeviceID, Host: peer.IP, Port: peer.Port}
	req := sip.NewRequest(sip.MESSAGE, recipient)
	req.SetBody([]byte(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	from := &sip.FromHeader{Address: sip.Uri{User: s.cfg.ID, Host: s.cfg.Domain}}
	to := &sip.FromHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return s.client.WriteRequest(req)
}

// SendCatalogQuery sends a Catalog query to the device.
func (s *Server) SendCatalogQuery(peer Peer) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildCatalogQuery(peer.DeviceID, sn))
}

// SendDeviceInfoQuery sends a DeviceInfo query.
func (s *Server) SendDeviceInfoQuery(peer Peer) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildDeviceInfoQuery(peer.DeviceID, sn))
}

// SendDeviceStatusQuery sends a DeviceStatus query and returns SN + result channel.
func (s *Server) SendDeviceStatusQuery(peer Peer) (string, <-chan *manscdp.DeviceStatus) {
	sn := s.NextSN()
	ch := s.status.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildDeviceStatusQuery(peer.DeviceID, sn))
	return sn, ch
}

// CancelDeviceStatusQuery cancels a pending status waiter.
func (s *Server) CancelDeviceStatusQuery(sn string) { s.status.Cancel(sn) }

// SendDeviceControl sends a raw MANSCDP control/query body.
func (s *Server) SendDeviceControl(peer Peer, xmlBody string) error {
	return s.sendMessage(peer, xmlBody)
}

// SendPTZ sends a direction PTZ command to channelID.
func (s *Server) SendPTZ(peer Peer, channelID, direction string, h, v, z int) error {
	sn := s.NextSN()
	cmd := manscdp.BuildDeviceControlPTZ(channelID, sn, ptz.Command(direction, h, v, z))
	return s.sendMessage(peer, cmd)
}

// SendFrontEndCmd sends a front-end PTZCmd (preset set/call/delete, etc.).
func (s *Server) SendFrontEndCmd(peer Peer, channelID string, cmdCode, parameter1, parameter2, combineCode2 int) error {
	sn := s.NextSN()
	cmd := ptz.FrontEndCmd(cmdCode, parameter1, parameter2, combineCode2)
	body := manscdp.BuildDeviceControlFrontEnd(channelID, sn, cmd)
	return s.sendMessage(peer, body)
}

// SendRecordCmd sends Record / StopRecord and optionally waits for Response via returned channel.
func (s *Server) SendRecordCmd(peer Peer, channelID, recordCmd string) (string, <-chan *session.ControlResult) {
	sn := s.NextSN()
	ch := s.controls.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildRecordCmd(channelID, sn, recordCmd))
	return sn, ch
}

// SendGuardCmd sends SetGuard / ResetGuard.
func (s *Server) SendGuardCmd(peer Peer, channelID, guardCmd string) (string, <-chan *session.ControlResult) {
	sn := s.NextSN()
	ch := s.controls.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildGuardCmd(channelID, sn, guardCmd))
	return sn, ch
}

// SendTeleBoot sends TeleBoot.
func (s *Server) SendTeleBoot(peer Peer, channelID string) (string, <-chan *session.ControlResult) {
	sn := s.NextSN()
	ch := s.controls.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildTeleBoot(channelID, sn))
	return sn, ch
}

// SendIFrameCmd sends IFrameCmd.
func (s *Server) SendIFrameCmd(peer Peer, channelID string) (string, <-chan *session.ControlResult) {
	sn := s.NextSN()
	ch := s.controls.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildIFrameCmd(channelID, sn))
	return sn, ch
}

// SendAlarmCmd sends AlarmCmd (e.g. ResetAlarm).
func (s *Server) SendAlarmCmd(peer Peer, channelID, alarmCmd string) (string, <-chan *session.ControlResult) {
	sn := s.NextSN()
	ch := s.controls.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildAlarmCmd(channelID, sn, alarmCmd))
	return sn, ch
}

// CancelControl cancels a pending DeviceControl waiter.
func (s *Server) CancelControl(sn string) { s.controls.Cancel(sn) }

// SendConfigDownloadQuery sends ConfigDownload.
func (s *Server) SendConfigDownloadQuery(peer Peer, configType string) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildConfigDownloadQuery(peer.DeviceID, sn, configType))
}

// SendHomePositionQuery sends HomePositionQuery.
func (s *Server) SendHomePositionQuery(peer Peer, channelID string) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildHomePositionQuery(channelID, sn))
}

// SendCruiseTrackListQuery sends CruiseTrackListQuery.
func (s *Server) SendCruiseTrackListQuery(peer Peer, channelID string) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildCruiseTrackListQuery(channelID, sn))
}

// SendCruiseTrackQuery sends CruiseTrackQuery.
func (s *Server) SendCruiseTrackQuery(peer Peer, channelID string, number int) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildCruiseTrackQuery(channelID, sn, number))
}

// SendPTZPositionQuery sends PTZPosition query.
func (s *Server) SendPTZPositionQuery(peer Peer, channelID string) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildPTZPositionQuery(channelID, sn))
}

// SendAudioBroadcast sends a Broadcast notify.
func (s *Server) SendAudioBroadcast(peer Peer, channelGBID string) error {
	sn := s.NextSN()
	return s.sendMessage(peer, manscdp.BuildBroadcastNotify(s.cfg.ID, channelGBID, sn))
}

// SendSubscribe sends a SIP SUBSCRIBE.
func (s *Server) SendSubscribe(peer Peer, eventType, body string, expiresSec int) error {
	if err := s.requirePeerAddr(peer); err != nil {
		return err
	}
	if expiresSec > 0 && expiresSec < 30 {
		expiresSec = 30
	}
	recipient := sip.Uri{User: peer.DeviceID, Host: peer.IP, Port: peer.Port}
	req := sip.NewRequest(sip.SUBSCRIBE, recipient)
	req.SetBody([]byte(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	req.AppendHeader(sip.NewHeader("Event", eventType))
	req.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(expiresSec)))
	from := &sip.FromHeader{Address: sip.Uri{User: s.cfg.ID, Host: s.cfg.Domain}}
	to := &sip.FromHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return s.client.WriteRequest(req)
}

// SendSubscribeCancel cancels a subscription (Expires=0).
func (s *Server) SendSubscribeCancel(peer Peer, eventType, body string) error {
	return s.SendSubscribe(peer, eventType, body, 0)
}

// SendRecordInfoQuery starts a RecordInfo query and returns SN + result channel.
func (s *Server) SendRecordInfoQuery(peer Peer, channelID, startTime, endTime string) (string, <-chan *session.RecordInfo) {
	return s.SendRecordInfoQueryOpts(peer, channelID, manscdp.RecordInfoOpts{
		StartTime: startTime,
		EndTime:   endTime,
		Type:      "all",
	})
}

// SendRecordInfoQueryOpts starts a RecordInfo query with optional RecLocation/RecordPos.
func (s *Server) SendRecordInfoQueryOpts(peer Peer, channelID string, opts manscdp.RecordInfoOpts) (string, <-chan *session.RecordInfo) {
	sn := s.NextSN()
	ch := s.records.Register(sn)
	body := manscdp.BuildRecordInfoQueryOpts(channelID, sn, opts)
	_ = s.sendMessage(peer, body)
	return sn, ch
}

// CancelRecordQuery cancels a pending record waiter.
func (s *Server) CancelRecordQuery(sn string) { s.records.Cancel(sn) }

// SendPresetQuery starts a PresetQuery and returns SN + result channel.
func (s *Server) SendPresetQuery(peer Peer, channelID string) (string, <-chan []manscdp.Preset) {
	sn := s.NextSN()
	ch := s.presets.Register(sn)
	_ = s.sendMessage(peer, manscdp.BuildPresetQuery(channelID, sn))
	return sn, ch
}

// CancelPresetQuery cancels a pending preset waiter.
func (s *Server) CancelPresetQuery(sn string) { s.presets.Cancel(sn) }
