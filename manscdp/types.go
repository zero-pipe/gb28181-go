package manscdp

// Message is a parsed MANSCDP XML document (Notify / Response / Query / Control).
type Message struct {
	Root         string
	CmdType      string
	SN           string
	DeviceID     string
	SumNum       int
	Result       string // Response Result (OK/ERROR)
	NotifyType   string // MediaStatus NotifyType
	Raw          []byte
	Items        []CatalogItem
	RecordItems  []RecordItem
	PresetItems  []Preset
	Alarm        *AlarmNotify
	Position     *MobilePositionNotify
	DeviceStatus *DeviceStatus
	MediaStatus  *MediaStatusNotify
	PTZCmd       string // DeviceControl PTZCmd (inbound from superior)
}

// CatalogItem is a device/channel entry in a Catalog response or notify.
type CatalogItem struct {
	DeviceID     string
	Name         string
	Manufacturer string
	Model        string
	Owner        string
	CivilCode    string
	Block        string
	Address      string
	Parental     int
	ParentID     string
	RegisterWay  int
	Secrecy      int
	Status       string
	Longitude    float64
	Latitude     float64
	PTZType      int
	Event        string // ON/OFF/ADD/DEL/UPDATE/VLOST/DEFECT (附录 J/N)
	OperateType  string // vendor (e.g. Hik) alternate of Event
}

// RecordItem is a record file entry in a RecordInfo response.
type RecordItem struct {
	DeviceID   string
	Name       string
	FilePath   string
	FileSize   string
	StartTime  string
	EndTime    string
	Secrecy    int
	Type       string
	RecorderID string
}

// Preset is a PTZ preset from PresetQuery.
type Preset struct {
	PresetID   string
	PresetName string
}

// AlarmNotify is an Alarm MANSCDP notify payload.
type AlarmNotify struct {
	AlarmPriority    string
	AlarmMethod      string
	AlarmTime        string
	AlarmDescription string
	Longitude        float64
	Latitude         float64
	AlarmType        int
}

// MobilePositionNotify is a MobilePosition MANSCDP notify payload.
type MobilePositionNotify struct {
	Longitude float64
	Latitude  float64
	Speed     float64
	Direction float64
	Altitude  float64
	Time      string
}

// DeviceStatus is a DeviceStatus query response.
type DeviceStatus struct {
	Result      string
	Online      string // ONLINE / OFFLINE
	Status      string // OK / ERROR
	Encode      string // ON / OFF
	Record      string // ON / OFF (国标字段名 Record)
	DeviceTime  string
	AlarmStatus string // raw Alarmstatus block summary
}

// MediaStatusNotify is a MediaStatus notify (e.g. download end NotifyType=121).
type MediaStatusNotify struct {
	NotifyType string
	DeviceID   string
}

// RecordInfoOpts configures a RecordInfo query.
type RecordInfoOpts struct {
	StartTime   string
	EndTime     string
	Secrecy     int
	Type        string // all / time / alarm / manual …
	RecLocation string // optional, vendor/common extension
	RecordPos   string // optional
}
