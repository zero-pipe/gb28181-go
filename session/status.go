package session

import (
	"errors"
	"sync"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

var ErrStatusTimeout = errors.New("device status query timeout")

// StatusWaiter waits for DeviceStatus Response by SN.
type StatusWaiter struct {
	pending sync.Map
}

func NewStatusWaiter() *StatusWaiter {
	return &StatusWaiter{}
}

func (m *StatusWaiter) Register(sn string) <-chan *manscdp.DeviceStatus {
	ch := make(chan *manscdp.DeviceStatus, 1)
	m.pending.Store(sn, ch)
	return ch
}

func (m *StatusWaiter) Cancel(sn string) {
	if v, ok := m.pending.LoadAndDelete(sn); ok {
		close(v.(chan *manscdp.DeviceStatus))
	}
}

func (m *StatusWaiter) Handle(sn string, st *manscdp.DeviceStatus) {
	v, ok := m.pending.LoadAndDelete(sn)
	if !ok {
		return
	}
	ch := v.(chan *manscdp.DeviceStatus)
	select {
	case ch <- st:
	default:
	}
}

// Wait blocks on the channel returned by Register. See RecordWaiter.Wait for
// why the channel is passed explicitly.
func (m *StatusWaiter) Wait(sn string, ch <-chan *manscdp.DeviceStatus, timeout time.Duration) (*manscdp.DeviceStatus, error) {
	select {
	case st, ok := <-ch:
		if !ok {
			return nil, ErrStatusTimeout
		}
		return st, nil
	case <-time.After(timeout):
		m.Cancel(sn)
		return nil, ErrStatusTimeout
	}
}
