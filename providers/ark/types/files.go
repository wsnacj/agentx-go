package types

import "io"

// UploadFileRequest holds multipart upload data.
type UploadFileRequest struct {
	File              io.Reader          `json:"-"`
	Filename          string             `json:"-"`
	Purpose           string             `json:"purpose"`
	ExpireAt          *int64             `json:"expire_at,omitempty"`
	PreprocessConfigs *PreprocessConfigs `json:"preprocess_configs,omitempty"`
}

// PreprocessConfigs defines preprocessing rules for files.
type PreprocessConfigs struct {
	Video *VideoPreprocess `json:"video,omitempty"`
}

// VideoPreprocess describes video preprocessing settings.
type VideoPreprocess struct {
	Fps   *float64 `json:"fps,omitempty"`
	Model *string  `json:"model,omitempty"`
}

// FileObject represents a file object.
type FileObject struct {
	Object           string             `json:"object,omitempty"`
	ID               string             `json:"id,omitempty"`
	Purpose          string             `json:"purpose,omitempty"`
	Filename         string             `json:"filename,omitempty"`
	Bytes            int64              `json:"bytes,omitempty"`
	MimeType         string             `json:"mime_type,omitempty"`
	CreatedAt        int64              `json:"created_at,omitempty"`
	ExpireAt         int64              `json:"expire_at,omitempty"`
	Status           string             `json:"status,omitempty"`
	Error            *APIError          `json:"error,omitempty"`
	PreprocessConfig *PreprocessConfigs `json:"preprocess_config,omitempty"`
}

// FileList represents a list response.
type FileList struct {
	Object  string       `json:"object,omitempty"`
	Data    []FileObject `json:"data,omitempty"`
	FirstID string       `json:"first_id,omitempty"`
	LastID  string       `json:"last_id,omitempty"`
	HasMore bool         `json:"has_more,omitempty"`
}

// ListFileOptions defines list query parameters.
type ListFileOptions struct {
	After   string
	Limit   int
	Purpose string
	Order   string
}

// ListInputOptions defines input item listing query parameters.
type ListInputOptions struct {
	After   string
	Before  string
	Limit   int
	Order   string
	Include []string
}

// InputItemList is used for listing inputs.
type InputItemList struct {
	Object  string      `json:"object,omitempty"`
	Data    []InputItem `json:"data,omitempty"`
	FirstID string      `json:"first_id,omitempty"`
	LastID  string      `json:"last_id,omitempty"`
	HasMore bool        `json:"has_more,omitempty"`
}
