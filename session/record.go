package session

import (
	"errors"
	"sync"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

var ErrRecordTimeout = errors.New("record query timeout")

// RecordInfo is an aggregated RecordInfo response.
type RecordInfo struct {
	DeviceID   string
	ChannelID  string
	SN         string
	SumNum     int
	Count      int
	RecordList []manscdp.RecordItem
}

// RecordWaiter aggregates fragmented RecordInfo responses by SN.
type RecordWaiter struct {
	pending sync.Map
}

func NewRecordWaiter() *RecordWaiter {
	return &RecordWaiter{}
}

type recordWaiter struct {
	ch     chan *RecordInfo
	items  []manscdp.RecordItem
	sumNum int
	sn     string
}

func (m *RecordWaiter) Register(sn string) <-chan *RecordInfo {
	w := &recordWaiter{ch: make(chan *RecordInfo, 1), sn: sn}
	m.pending.Store(sn, w)
	return w.ch
}

func (m *RecordWaiter) Cancel(sn string) {
	if v, ok := m.pending.LoadAndDelete(sn); ok {
		close(v.(*recordWaiter).ch)
	}
}

func (m *RecordWaiter) Wait(sn string, timeout time.Duration) (*RecordInfo, error) {
	v, ok := m.pending.Load(sn)
	if !ok {
		ch := m.Register(sn)
		select {
		case info := <-ch:
			m.pending.Delete(sn)
			return info, nil
		case <-time.After(timeout):
			m.Cancel(sn)
			return nil, ErrRecordTimeout
		}
	}
	w := v.(*recordWaiter)
	select {
	case info := <-w.ch:
		m.pending.Delete(sn)
		return info, nil
	case <-time.After(timeout):
		m.Cancel(sn)
		return nil, ErrRecordTimeout
	}
}

func (m *RecordWaiter) Handle(deviceID, channelID, sn string, sumNum int, items []manscdp.RecordItem) {
	v, ok := m.pending.Load(sn)
	if !ok {
		return
	}
	w := v.(*recordWaiter)
	w.sumNum = sumNum
	w.items = append(w.items, items...)
	if sumNum == 0 || len(w.items) >= sumNum {
		info := &RecordInfo{
			DeviceID: deviceID, ChannelID: channelID, SN: sn,
			SumNum: sumNum, Count: len(w.items),
			RecordList: append([]manscdp.RecordItem(nil), w.items...),
		}
		select {
		case w.ch <- info:
		default:
		}
		m.pending.Delete(sn)
	}
}
