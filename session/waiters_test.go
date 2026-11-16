package session

import (
	"testing"
	"time"

	"github.com/zero-pipe/gb28181-go/manscdp"
)

func TestRecordWaiterDeliversBeforeWait(t *testing.T) {
	m := NewRecordWaiter()
	ch := m.Register("sn-1")
	// Response arrives before Wait is called; previously this raced to a timeout.
	m.Handle("dev", "ch", "sn-1", 1, []manscdp.RecordItem{{Name: "a"}})

	info, err := m.Wait("sn-1", ch, time.Second)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if info == nil || info.Count != 1 || len(info.RecordList) != 1 {
		t.Fatalf("unexpected record info: %#v", info)
	}
}

func TestRecordWaiterAggregatesFragments(t *testing.T) {
	m := NewRecordWaiter()
	ch := m.Register("sn-2")
	m.Handle("dev", "ch", "sn-2", 2, []manscdp.RecordItem{{Name: "a"}})
	m.Handle("dev", "ch", "sn-2", 2, []manscdp.RecordItem{{Name: "b"}})

	info, err := m.Wait("sn-2", ch, time.Second)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if info.Count != 2 {
		t.Fatalf("expected 2 items, got %d", info.Count)
	}
}

func TestRecordWaiterTimeoutCancels(t *testing.T) {
	m := NewRecordWaiter()
	ch := m.Register("sn-3")
	if _, err := m.Wait("sn-3", ch, 10*time.Millisecond); err != ErrRecordTimeout {
		t.Fatalf("expected timeout, got %v", err)
	}
	// Cancel closed the channel.
	if _, ok := <-ch; ok {
		t.Fatal("expected channel closed after timeout")
	}
}

func TestPresetWaiterDeliversBeforeWait(t *testing.T) {
	m := NewPresetWaiter()
	ch := m.Register("sn-p")
	m.Handle("sn-p", 1, []manscdp.Preset{{PresetID: "1", PresetName: "p1"}})

	items, err := m.Wait("sn-p", ch, time.Second)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(items) != 1 || items[0].PresetName != "p1" {
		t.Fatalf("unexpected presets: %#v", items)
	}
}

func TestStatusWaiterDeliversBeforeWait(t *testing.T) {
	m := NewStatusWaiter()
	ch := m.Register("sn-s")
	m.Handle("sn-s", &manscdp.DeviceStatus{Online: "ONLINE"})

	st, err := m.Wait("sn-s", ch, time.Second)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if st == nil || st.Online != "ONLINE" {
		t.Fatalf("unexpected status: %#v", st)
	}
}

func TestControlWaiterDeliversBeforeWait(t *testing.T) {
	m := NewControlWaiter()
	ch := m.Register("sn-c")
	m.Handle("sn-c", "dev", "OK")

	r, err := m.Wait("sn-c", ch, time.Second)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if r.Result != "OK" || r.DeviceID != "dev" {
		t.Fatalf("unexpected control result: %#v", r)
	}
}
