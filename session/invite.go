package session

import (
	"errors"
	"sync"
	"time"

	"github.com/emiago/sipgo"
)

type SessionType string

const (
	SessionPlay     SessionType = "PLAY"
	SessionPlayback SessionType = "PLAYBACK"
	SessionDownload SessionType = "DOWNLOAD"
)

var ErrSessionNotFound = errors.New("invite session not found")

// InviteSession holds an active INVITE dialog.
type InviteSession struct {
	DeviceID      string
	ChannelID     string
	IP            string
	Port          int
	Stream        string
	App           string
	Type          SessionType
	Dialog        *sipgo.DialogClientSession
	StartTime     string
	EndTime       string
	DownloadSpeed int
	StartedAt     time.Time
}

// InviteManager stores invite sessions by stream id.
type InviteManager struct {
	sessions sync.Map
}

func NewInviteManager() *InviteManager {
	return &InviteManager{}
}

func (m *InviteManager) Put(stream string, sess *InviteSession) {
	m.sessions.Store(stream, sess)
}

func (m *InviteManager) Get(stream string) (*InviteSession, bool) {
	v, ok := m.sessions.Load(stream)
	if !ok {
		return nil, false
	}
	return v.(*InviteSession), true
}

func (m *InviteManager) Remove(stream string) {
	m.sessions.Delete(stream)
}

// RemoveByCallID removes sessions whose INVITE Call-ID matches and returns stream keys.
func (m *InviteManager) RemoveByCallID(callID string) []string {
	if callID == "" {
		return nil
	}
	var removed []string
	m.sessions.Range(func(k, v any) bool {
		sess := v.(*InviteSession)
		if sess == nil || sess.Dialog == nil || sess.Dialog.InviteRequest == nil {
			return true
		}
		cid := sess.Dialog.InviteRequest.CallID()
		if cid == nil {
			return true
		}
		if cid.Value() == callID {
			m.sessions.Delete(k)
			removed = append(removed, k.(string))
		}
		return true
	})
	return removed
}

func (s *InviteSession) Progress() float64 {
	if s.Type != SessionDownload || s.StartTime == "" || s.EndTime == "" {
		return 0
	}
	const layout = "2006-01-02 15:04:05"
	start, err1 := time.ParseInLocation(layout, s.StartTime, time.Local)
	end, err2 := time.ParseInLocation(layout, s.EndTime, time.Local)
	if err1 != nil || err2 != nil || !end.After(start) {
		return 0
	}
	total := end.Sub(start).Seconds()
	elapsed := time.Since(s.StartedAt).Seconds()
	if downloadSpeed := s.DownloadSpeed; downloadSpeed > 0 {
		elapsed *= float64(downloadSpeed)
	}
	p := elapsed / total * 100
	if p > 100 {
		p = 100
	}
	if p < 0 {
		p = 0
	}
	return p
}
