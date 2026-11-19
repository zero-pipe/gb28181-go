package session

import (
	"errors"
	"sync"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

// ErrQueryTimeout is returned when no response arrives in time.
var ErrQueryTimeout = errors.New("query response timeout")

// QueryWaiter waits for a query Response (ConfigDownload, and the 2022 query
// family) keyed by SN. The whole parsed Message is delivered so callers can
// pick the payload they expect.
type QueryWaiter struct {
	pending sync.Map
}

func NewQueryWaiter() *QueryWaiter {
	return &QueryWaiter{}
}

func (m *QueryWaiter) Register(sn string) <-chan *manscdp.Message {
	ch := make(chan *manscdp.Message, 1)
	m.pending.Store(sn, ch)
	return ch
}

func (m *QueryWaiter) Cancel(sn string) {
	if v, ok := m.pending.LoadAndDelete(sn); ok {
		close(v.(chan *manscdp.Message))
	}
}

func (m *QueryWaiter) Handle(sn string, msg *manscdp.Message) bool {
	v, ok := m.pending.LoadAndDelete(sn)
	if !ok {
		return false
	}
	ch := v.(chan *manscdp.Message)
	select {
	case ch <- msg:
	default:
	}
	return true
}

// Wait blocks on the channel returned by Register.
func (m *QueryWaiter) Wait(sn string, ch <-chan *manscdp.Message, timeout time.Duration) (*manscdp.Message, error) {
	select {
	case msg, ok := <-ch:
		if !ok {
			return nil, ErrQueryTimeout
		}
		return msg, nil
	case <-time.After(timeout):
		m.Cancel(sn)
		return nil, ErrQueryTimeout
	}
}
