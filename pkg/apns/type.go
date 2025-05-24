package apns

// D provide string array
type D map[string]interface{}

const (
	// ApnsPriorityLow will tell APNs to send the push message at a time that takes
	// into account power considerations for the device. Notifications with this
	// priority might be grouped and delivered in bursts. They are throttled, and
	// in some cases are not delivered.
	ApnsPriorityLow = 5

	// ApnsPriorityHigh will tell APNs to send the push message immediately.
	// Notifications with this priority must trigger an alert, sound, or badge on
	// the target device. It is an error to use this priority for a push
	// notification that contains only the content-available key.
	ApnsPriorityHigh = 10
)

// Alert is APNs payload
type Alert struct {
	Action          string   `json:"action,omitempty"`
	ActionLocKey    string   `json:"action-loc-key,omitempty"`
	Body            string   `json:"body,omitempty"`
	LaunchImage     string   `json:"launch-image,omitempty"`
	LocArgs         []string `json:"loc-args,omitempty"`
	LocKey          string   `json:"loc-key,omitempty"`
	Title           string   `json:"title,omitempty"`
	Subtitle        string   `json:"subtitle,omitempty"`
	TitleLocArgs    []string `json:"title-loc-args,omitempty"`
	TitleLocKey     string   `json:"title-loc-key,omitempty"`
	SummaryArg      string   `json:"summary-arg,omitempty"`
	SummaryArgCount int      `json:"summary-arg-count,omitempty"`
}

// RequestPush support multiple notification request.
// type RequestPush struct {
// 	Notifications []PushNotification `json:"notifications" binding:"required"`
// }

// PushNotification is single notification request
type PushNotification struct {
	// Common
	ID               string      `json:"notif_id,omitempty"`
	To               string      `json:"to,omitempty"`
	Topic            string      `json:"topic,omitempty"` // FCM and iOS only
	Tokens           []string    `json:"tokens" binding:"required"`
	Platform         int         `json:"platform" binding:"required"`
	Message          string      `json:"message,omitempty"`
	Title            string      `json:"title,omitempty"`
	Image            string      `json:"image,omitempty"`
	Priority         string      `json:"priority,omitempty"`
	ContentAvailable bool        `json:"content_available,omitempty"`
	MutableContent   bool        `json:"mutable_content,omitempty"`
	Sound            interface{} `json:"sound,omitempty"`
	Data             D           `json:"data,omitempty"`
	Retry            int         `json:"retry,omitempty"`

	// iOS
	Expiration  *int64   `json:"expiration,omitempty"`
	ApnsID      string   `json:"apns_id,omitempty"`
	CollapseID  string   `json:"collapse_id,omitempty"`
	PushType    string   `json:"push_type,omitempty"`
	Badge       *int     `json:"badge,omitempty"`
	Category    string   `json:"category,omitempty"`
	ThreadID    string   `json:"thread-id,omitempty"`
	URLArgs     []string `json:"url-args,omitempty"`
	Alert       Alert    `json:"alert,omitempty"`
	Production  bool     `json:"production,omitempty"`
	Development bool     `json:"development,omitempty"`
	SoundName   string   `json:"name,omitempty"`
	SoundVolume float32  `json:"volume,omitempty"`

	// ref: https://github.com/sideshow/apns2/blob/54928d6193dfe300b6b88dad72b7e2ae138d4f0a/payload/builder.go#L7-L24
	InterruptionLevel string `json:"interruption_level,omitempty"`
}
