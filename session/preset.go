package session

import (
	"errors"
	"sync"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

var ErrPresetTimeout = errors.New("预置位查询超时，设备未返回")

// PresetWaiter aggregates PresetQuery responses by SN.
type PresetWaiter struct {
	pending sync.Map
}

func NewPresetWaiter() *PresetWaiter {
	return &PresetWaiter{}
}

type presetWaiter struct {
	ch     chan []manscdp.Preset
	items  []manscdp.Preset
	sumNum int
	sn     string
}

func (m *PresetWaiter) Register(sn string) <-chan []manscdp.Preset {
	w := &presetWaiter{ch: make(chan []manscdp.Preset, 1), sn: sn}
	m.pending.Store(sn, w)
	return w.ch
}

func (m *PresetWaiter) Cancel(sn string) {
	if v, ok := m.pending.LoadAndDelete(sn); ok {
		close(v.(*presetWaiter).ch)
	}
}

func (m *PresetWaiter) Handle(sn string, sumNum int, items []manscdp.Preset) {
	v, ok := m.pending.Load(sn)
	if !ok {
		return
	}
	w := v.(*presetWaiter)
	if sumNum > 0 {
		w.sumNum = sumNum
	}
	w.items = append(w.items, items...)
	target := w.sumNum
	if target == 0 && len(items) == 0 {
		select {
		case w.ch <- []manscdp.Preset{}:
		default:
		}
		m.pending.Delete(sn)
		return
	}
	if target == 0 || len(w.items) >= target {
		out := append([]manscdp.Preset(nil), w.items...)
		select {
		case w.ch <- out:
		default:
		}
		m.pending.Delete(sn)
	}
}

func (m *PresetWaiter) Wait(sn string, timeout time.Duration) ([]manscdp.Preset, error) {
	v, ok := m.pending.Load(sn)
	if !ok {
		return nil, ErrPresetTimeout
	}
	w := v.(*presetWaiter)
	select {
	case info := <-w.ch:
		m.pending.Delete(sn)
		if info == nil {
			return []manscdp.Preset{}, nil
		}
		return info, nil
	case <-time.After(timeout):
		m.Cancel(sn)
		return nil, ErrPresetTimeout
	}
}
