package types

import (
	"encoding/json"
	"fmt"
)

// Event is a raw SSE event with its type and payload.
type Event struct {
	Type           string          `json:"type"`
	SequenceNumber int             `json:"sequence_number,omitempty"`
	Raw            json.RawMessage `json:"-"`
}

const (
	EventTypeResponseCreated                   = "response.created"
	EventTypeResponseInProgress                = "response.in_progress"
	EventTypeResponseCompleted                 = "response.completed"
	EventTypeResponseFailed                    = "response.failed"
	EventTypeResponseIncomplete                = "response.incomplete"
	EventTypeResponseOutputItemAdded           = "response.output_item.added"
	EventTypeResponseOutputItemDone            = "response.output_item.done"
	EventTypeResponseContentPartAdded          = "response.content_part.added"
	EventTypeResponseContentPartDone           = "response.content_part.done"
	EventTypeResponseOutputTextDelta           = "response.output_text.delta"
	EventTypeResponseOutputTextDone            = "response.output_text.done"
	EventTypeResponseFunctionCallArgsDelta     = "response.function_call_arguments.delta"
	EventTypeResponseFunctionCallArgsDone      = "response.function_call_arguments.done"
	EventTypeResponseReasoningSummaryAdded     = "response.reasoning_summary_part.added"
	EventTypeResponseReasoningSummaryDone      = "response.reasoning_summary_part.done"
	EventTypeResponseReasoningSummaryTextDelta = "response.reasoning_summary_text.delta"
	EventTypeResponseReasoningSummaryTextDone  = "response.reasoning_summary_text.done"
	EventTypeError                             = "error"
)

// DecodeEvent decodes a raw SSE payload into an Event.
func DecodeEvent(data []byte) (Event, error) {
	var meta struct {
		Type           string `json:"type"`
		SequenceNumber int    `json:"sequence_number"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return Event{}, err
	}
	return Event{Type: meta.Type, SequenceNumber: meta.SequenceNumber, Raw: data}, nil
}

// DecodeTypedEvent decodes an Event payload into a typed struct.
func DecodeTypedEvent(ev Event) (any, error) {
	switch ev.Type {
	case EventTypeResponseCreated, EventTypeResponseInProgress, EventTypeResponseCompleted, EventTypeResponseFailed, EventTypeResponseIncomplete:
		var out ResponseEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseOutputItemAdded, EventTypeResponseOutputItemDone:
		var out OutputItemEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseContentPartAdded, EventTypeResponseContentPartDone:
		var out ContentPartEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseOutputTextDelta:
		var out OutputTextDeltaEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseOutputTextDone:
		var out OutputTextDoneEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseFunctionCallArgsDelta:
		var out FunctionCallArgumentsDeltaEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseFunctionCallArgsDone:
		var out FunctionCallArgumentsDoneEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseReasoningSummaryTextDelta:
		var out ReasoningSummaryTextDeltaEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeResponseReasoningSummaryTextDone:
		var out ReasoningSummaryTextDoneEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case EventTypeError:
		var out ErrorEvent
		if err := json.Unmarshal(ev.Raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("ark types: unsupported event type %s", ev.Type)
	}
}

// StreamResult provides streaming events and lifecycle control.
type StreamResult struct {
	Ch     <-chan Event
	Err    <-chan error
	Cancel func()
}

// ResponseEvent carries a response object.
type ResponseEvent struct {
	Type           string    `json:"type"`
	SequenceNumber int       `json:"sequence_number,omitempty"`
	Response       *Response `json:"response,omitempty"`
}

// OutputItemEvent represents events with an output item.
type OutputItemEvent struct {
	Type           string      `json:"type"`
	SequenceNumber int         `json:"sequence_number,omitempty"`
	OutputIndex    int         `json:"output_index,omitempty"`
	Item           *OutputItem `json:"item,omitempty"`
}

// ContentPartEvent represents content part events.
type ContentPartEvent struct {
	Type           string          `json:"type"`
	SequenceNumber int             `json:"sequence_number,omitempty"`
	OutputIndex    int             `json:"output_index,omitempty"`
	ItemID         string          `json:"item_id,omitempty"`
	ContentIndex   int             `json:"content_index,omitempty"`
	Part           json.RawMessage `json:"part,omitempty"`
}

// OutputTextDeltaEvent represents output text deltas.
type OutputTextDeltaEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	ContentIndex   int    `json:"content_index,omitempty"`
	Delta          string `json:"delta,omitempty"`
}

// OutputTextDoneEvent represents completed output text.
type OutputTextDoneEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	ContentIndex   int    `json:"content_index,omitempty"`
	Text           string `json:"text,omitempty"`
}

// ReasoningSummaryTextDeltaEvent represents reasoning summary deltas.
type ReasoningSummaryTextDeltaEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	SummaryIndex   int    `json:"summary_index,omitempty"`
	Delta          string `json:"delta,omitempty"`
}

// ReasoningSummaryTextDoneEvent represents completed reasoning summary text.
type ReasoningSummaryTextDoneEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	SummaryIndex   int    `json:"summary_index,omitempty"`
	Text           string `json:"text,omitempty"`
}

// FunctionCallArgumentsDeltaEvent represents function call arguments deltas.
type FunctionCallArgumentsDeltaEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	Delta          string `json:"delta,omitempty"`
}

// FunctionCallArgumentsDoneEvent represents completed function call arguments.
type FunctionCallArgumentsDoneEvent struct {
	Type           string `json:"type"`
	SequenceNumber int    `json:"sequence_number,omitempty"`
	OutputIndex    int    `json:"output_index,omitempty"`
	ItemID         string `json:"item_id,omitempty"`
	Arguments      string `json:"arguments,omitempty"`
}

// ErrorEvent represents an error SSE event.
type ErrorEvent struct {
	Type           string    `json:"type"`
	SequenceNumber int       `json:"sequence_number,omitempty"`
	Code           string    `json:"code,omitempty"`
	Message        string    `json:"message,omitempty"`
	Param          string    `json:"param,omitempty"`
	Error          *APIError `json:"error,omitempty"`
}
