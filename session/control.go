package session

import (
	"errors"
	"sync"
	"time"
)

var ErrControlTimeout = errors.New("device control response timeout")

// ControlResult is a DeviceControl Response Result.
type ControlResult struct {
	SN       string
	DeviceID string
	Result   string
}

// ControlWaiter waits for DeviceControl Response by SN.
type ControlWaiter struct {
	pending sync.Map
}

func NewControlWaiter() *ControlWaiter {
	return &ControlWaiter{}
}

func (m *ControlWaiter) Register(sn string) <-chan *ControlResult {
	ch := make(chan *ControlResult, 1)
	m.pending.Store(sn, ch)
	return ch
}

func (m *ControlWaiter) Cancel(sn string) {
	if v, ok := m.pending.LoadAndDelete(sn); ok {
		close(v.(chan *ControlResult))
	}
}

func (m *ControlWaiter) Handle(sn, deviceID, result string) {
	v, ok := m.pending.LoadAndDelete(sn)
	if !ok {
		return
	}
	ch := v.(chan *ControlResult)
	select {
	case ch <- &ControlResult{SN: sn, DeviceID: deviceID, Result: result}:
	default:
	}
}

func (m *ControlWaiter) Wait(sn string, timeout time.Duration) (*ControlResult, error) {
	v, ok := m.pending.Load(sn)
	if !ok {
		return nil, ErrControlTimeout
	}
	ch := v.(chan *ControlResult)
	select {
	case r := <-ch:
		m.pending.Delete(sn)
		return r, nil
	case <-time.After(timeout):
		m.Cancel(sn)
		return nil, ErrControlTimeout
	}
}
