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

	// 2022 / ConfigDownload 查询应答。
	BasicParam      *BasicParam
	HomePosition    *HomePosition
	CruiseTrackList *CruiseTrackList
	CruiseTrack     *CruiseTrack
	PTZPosition     *PTZPosition
	SDCardStatus    *SDCardStatus

	// 图像抓拍（2022 9.14）。
	SnapShot *SnapShotFinished
}

// SnapShotFinished is an UploadSnapShotFinished notify (A.2.5.7).
type SnapShotFinished struct {
	SessionID string
	Images    []SnapShotImage
}

// SnapShotImage is one uploaded image reference.
type SnapShotImage struct {
	ImageID  string
	FilePath string
}

// BasicParam is a ConfigDownload(BasicParam) response.
type BasicParam struct {
	Result             string
	Name               string
	Expiration         string
	HeartBeatInterval  string
	HeartBeatCount     string
	PositionCapability string
	Longitude          string
	Latitude           string
}

// HomePosition is a HomePositionQuery response (2022).
type HomePosition struct {
	Result      string
	Enabled     int
	ResetTime   int
	PresetIndex int
}

// CruisePoint is one stop in a cruise track.
type CruisePoint struct {
	PointID  int
	PresetID int
	Name     string
	StayTime int
	Speed    int
}

// CruiseTrack is one cruise track.
type CruiseTrack struct {
	Number int
	Name   string
	Points []CruisePoint
}

// CruiseTrackList is a CruiseTrackListQuery response.
type CruiseTrackList struct {
	Result string
	Num    int
	Tracks []CruiseTrack
}

// PTZPosition is a PTZPosition query/notify payload (2022).
type PTZPosition struct {
	Result string
	Pan    int
	Tilt   int
	Zoom   int
}

// SDCardStatus is an SDCardStatus query response (2022).
type SDCardStatus struct {
	Result     string
	Status     string
	Type       string
	CapacityMB int
	FreeMB     int
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
	// 2022 AlarmTypeParam 扩展字段。
	AlarmTypeParam string
	EventType      string
	EventData      string
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
