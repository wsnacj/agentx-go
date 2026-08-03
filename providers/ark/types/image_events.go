package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ImageGenerationEvent is a raw SSE event for Images API.
type ImageGenerationEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

const (
	EventTypeImageGenerationPartialSucceeded = "image_generation.partial_succeeded"
	EventTypeImageGenerationPartialFailed    = "image_generation.partial_failed"
	EventTypeImageGenerationCompleted        = "image_generation.completed"
	EventTypeImageGenerationError            = "error"
)

// DecodeImageGenerationEvent decodes a raw SSE payload into an image event.
func DecodeImageGenerationEvent(data []byte) (ImageGenerationEvent, error) {
	var meta struct {
		Type  string    `json:"type"`
		Error *APIError `json:"error"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return ImageGenerationEvent{}, err
	}
	eventType := strings.TrimSpace(meta.Type)
	if eventType == "" && meta.Error != nil {
		eventType = EventTypeImageGenerationError
	}
	if eventType == "" {
		return ImageGenerationEvent{}, fmt.Errorf("ark types: image event type is required")
	}
	return ImageGenerationEvent{Type: eventType, Raw: data}, nil
}

// DecodeTypedImageGenerationEvent decodes a raw event into a typed struct.
func DecodeTypedImageGenerationEvent(ev ImageGenerationEvent) (any, error) {
	switch ev.Type {
	case EventTypeImageGenerationPartialSucceeded:
		var out ImageGenerationPartialSucceededEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeImageGenerationPartialFailed:
		var out ImageGenerationPartialFailedEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeImageGenerationCompleted:
		var out ImageGenerationCompletedEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeImageGenerationError:
		var out ImageGenerationErrorEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("ark types: unsupported image event type %s", ev.Type)
	}
}

// ImageGenerationStreamResult provides streaming image events and lifecycle control.
type ImageGenerationStreamResult struct {
	Ch     <-chan ImageGenerationEvent
	Err    <-chan error
	Cancel func()
}

// ImageGenerationPartialSucceededEvent is emitted when one image is generated.
type ImageGenerationPartialSucceededEvent struct {
	Type       string `json:"type"`
	Model      string `json:"model,omitempty"`
	Created    int64  `json:"created,omitempty"`
	ImageIndex int    `json:"image_index,omitempty"`
	URL        string `json:"url,omitempty"`
	B64JSON    string `json:"b64_json,omitempty"`
	Size       string `json:"size,omitempty"`
}

// ImageGenerationPartialFailedEvent is emitted when one image fails.
type ImageGenerationPartialFailedEvent struct {
	Type       string    `json:"type"`
	Model      string    `json:"model,omitempty"`
	Created    int64     `json:"created,omitempty"`
	ImageIndex int       `json:"image_index,omitempty"`
	Error      *APIError `json:"error,omitempty"`
}

// ImageGenerationCompletedEvent is the last event in an image stream.
type ImageGenerationCompletedEvent struct {
	Type    string                `json:"type"`
	Model   string                `json:"model,omitempty"`
	Created int64                 `json:"created,omitempty"`
	Tools   []ImageGenerationTool `json:"tools,omitempty"`
	Usage   *ImageGenerationUsage `json:"usage,omitempty"`
}

// ImageGenerationErrorEvent carries request-level errors.
type ImageGenerationErrorEvent struct {
	Type    string    `json:"type,omitempty"`
	Code    string    `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
	Param   string    `json:"param,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}
