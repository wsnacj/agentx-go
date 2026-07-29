package mediaartifact

type Descriptor struct {
	Source        string `json:"source"`
	Kind          string `json:"kind"`
	Path          string `json:"path,omitempty"`
	URL           string `json:"url,omitempty"`
	MIMEType      string `json:"mime_type,omitempty"`
	Format        string `json:"format,omitempty"`
	Bytes         int64  `json:"bytes,omitempty"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	DurationMs    int64  `json:"duration_ms,omitempty"`
	TimestampMs   int64  `json:"timestamp_ms,omitempty"`
	FPS           int64  `json:"fps,omitempty"`
	HasAudio      *bool  `json:"has_audio,omitempty"`
	ScreenIndex   int64  `json:"screen_index,omitempty"`
	CaptureScope  string `json:"capture_scope,omitempty"`
	CaptureWidth  int    `json:"capture_width,omitempty"`
	CaptureHeight int    `json:"capture_height,omitempty"`
	Facing        string `json:"facing,omitempty"`
	Index         int    `json:"index,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
}
