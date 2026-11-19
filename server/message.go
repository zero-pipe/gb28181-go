package server

import (
	"context"
	"time"

	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/manscdp"
)

func (s *Server) handleMessage(req *sip.Request, tx sip.ServerTransaction) {
	fromDeviceID := extractSIPUser(req)
	body := req.Body()
	msg, err := manscdp.Parse(body)
	if err != nil {
		s.respond(tx, req, 400, "Bad Request")
		return
	}
	deviceID := fromDeviceID
	if deviceID == "" {
		deviceID = msg.DeviceID
	}

	// Superior → subordinate: Control/DeviceControl before DeviceKnown gate.
	if msg.Root == "Control" && msg.CmdType == "DeviceControl" &&
		s.handlers.CascadeInbound != nil && s.handlers.CascadeInbound.UpstreamKnown(deviceID) {
		channelID := msg.DeviceID
		ptz := msg.PTZCmd
		if ptz == "" {
			ptz = manscdp.ExtractTag(body, "PTZCmd")
		}
		_ = s.handlers.CascadeInbound.OnDeviceControl(context.Background(), InboundControlEvent{
			UpstreamGBID: deviceID,
			ChannelGBID:  channelID,
			SN:           msg.SN,
			PTZCmd:       ptz,
			Raw:          append([]byte(nil), body...),
		})
		s.respond(tx, req, 200, "OK")
		return
	}

	// Superior → subordinate: MANSCDP Query (Catalog/DeviceInfo/...) answered by us.
	if msg.Root == "Query" && s.handlers.CascadeInbound != nil && s.handlers.CascadeInbound.UpstreamKnown(deviceID) {
		respBody, ok := s.handlers.CascadeInbound.OnQuery(context.Background(), InboundQueryEvent{
			UpstreamGBID: deviceID,
			CmdType:      msg.CmdType,
			SN:           msg.SN,
			DeviceID:     msg.DeviceID,
			Raw:          append([]byte(nil), body...),
		})
		if !ok || len(respBody) == 0 {
			s.respond(tx, req, 200, "OK")
			return
		}
		res := sip.NewResponseFromRequest(req, 200, "OK", manscdp.EncodeBody(string(respBody)))
		res.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
		_ = tx.Respond(res)
		return
	}

	if s.handlers.Message == nil || !s.handlers.Message.DeviceKnown(deviceID) {
		s.respond(tx, req, 404, "Not Found")
		return
	}

	ctx := context.Background()
	switch msg.Root {
	case "Notify":
		switch msg.CmdType {
		case "Keepalive":
			ip, port := extractSourceAddr(req)
			_ = s.handlers.Message.OnKeepalive(ctx, deviceID, ip, port)
			if s.handlers.Telemetry != nil {
				s.handlers.Telemetry.OnKeepalive(deviceID, time.Now())
			}
		case "Catalog":
			s.logf("[gb28181-go] catalog notify: device=%s items=%d", deviceID, len(msg.Items))
			_ = s.handlers.Message.OnCatalog(ctx, deviceID, msg.Items, true)
		case "Alarm":
			_ = s.handlers.Message.OnAlarm(ctx, deviceID, msg.DeviceID, msg.Alarm)
		case "MobilePosition":
			_ = s.handlers.Message.OnMobilePosition(ctx, deviceID, msg.DeviceID, msg.Position)
		case "PTZPosition":
			_ = s.handlers.Message.OnPTZPosition(ctx, deviceID, msg.DeviceID, msg.PTZPosition)
		case "UploadSnapShotFinished":
			_ = s.handlers.Message.OnSnapShotFinished(ctx, deviceID, msg.SnapShot)
		case "MediaStatus":
			st := msg.MediaStatus
			if st == nil {
				st = &manscdp.MediaStatusNotify{NotifyType: msg.NotifyType, DeviceID: msg.DeviceID}
			}
			s.logf("[gb28181-go] MediaStatus: device=%s notifyType=%s", deviceID, st.NotifyType)
			_ = s.handlers.Message.OnMediaStatus(ctx, deviceID, st)
		}
	case "Response":
		switch msg.CmdType {
		case "Catalog":
			s.logf("[gb28181-go] catalog response: device=%s items=%d", deviceID, len(msg.Items))
			_ = s.handlers.Message.OnCatalog(ctx, deviceID, msg.Items, false)
		case "DeviceInfo":
			devName := manscdp.ExtractTag(body, "DeviceName")
			if devName == "" {
				devName = manscdp.ExtractTag(body, "Name")
			}
			mfr := manscdp.ExtractTag(body, "Manufacturer")
			model := manscdp.ExtractTag(body, "Model")
			fw := manscdp.ExtractTag(body, "Firmware")
			s.logf("[gb28181-go] DeviceInfo: device=%s name=%s manufacturer=%s model=%s",
				deviceID, devName, mfr, model)
			_ = s.handlers.Message.OnDeviceInfo(ctx, deviceID, devName, mfr, model, fw)
		case "DeviceStatus":
			st := msg.DeviceStatus
			if st == nil {
				st = &manscdp.DeviceStatus{Result: msg.Result}
			}
			s.status.Handle(msg.SN, st)
			_ = s.handlers.Message.OnDeviceStatus(ctx, deviceID, st)
		case "RecordInfo":
			s.records.Handle(deviceID, msg.DeviceID, msg.SN, msg.SumNum, msg.RecordItems)
		case "PresetQuery":
			s.presets.Handle(msg.SN, msg.SumNum, msg.PresetItems)
		case "ConfigDownload", "HomePositionQuery", "CruiseTrackListQuery",
			"CruiseTrackQuery", "PTZPosition", "SDCardStatus":
			s.queries.Handle(msg.SN, msg)
		case "DeviceControl":
			result := msg.Result
			if result == "" {
				result = manscdp.ExtractTag(body, "Result")
			}
			s.controls.Handle(msg.SN, msg.DeviceID, result)
			_ = s.handlers.Message.OnDeviceControlResult(ctx, deviceID, msg.SN, result)
		}
	}

	s.respond(tx, req, 200, "OK")
}
