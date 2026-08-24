package mcpmoqt

const (
	dataNSResources     = "resources"
	dataNSTools         = "tools"
	dataNSNotifications = "notifications"
)

type DataTrackType int

const (
	DataTrackResources DataTrackType = iota
	DataTrackTools
	DataTrackNotifications
)

func (t DataTrackType) String() string {
	switch t {
	case DataTrackResources:
		return dataNSResources
	case DataTrackTools:
		return dataNSTools
	case DataTrackNotifications:
		return dataNSNotifications
	default:
		return "unknown"
	}
}

func resourcesNamespace(sessionID string) []string {
	return []string{controlNS0, sessionID, dataNSResources}
}

func toolsNamespace(sessionID string) []string {
	return []string{controlNS0, sessionID, dataNSTools}
}

func notificationsNamespace(sessionID string) []string {
	return []string{controlNS0, sessionID, dataNSNotifications}
}

func dataTrackNamespace(sessionID string, trackType DataTrackType) []string {
	switch trackType {
	case DataTrackResources:
		return resourcesNamespace(sessionID)
	case DataTrackTools:
		return toolsNamespace(sessionID)
	case DataTrackNotifications:
		return notificationsNamespace(sessionID)
	default:
		return []string{controlNS0, sessionID, "unknown"}
	}
}

type DataTrackInfo struct {
	Type        DataTrackType
	TrackName   string
	Namespace   []string
	Description string
}

type DataTrackMessage struct {
	TrackType DataTrackType          `json:"track_type"`
	TrackName string                 `json:"track_name"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
