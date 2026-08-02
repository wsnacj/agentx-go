// Package contracts defines provider-neutral document data exchanged between
// OCR, PDF, pipeline and tool adapters.
package contracts

// BoundingBox identifies a rectangular region on one document page.
type BoundingBox struct {
	Page             int     `json:"page,omitempty"`
	X0               float64 `json:"x0"`
	Y0               float64 `json:"y0"`
	X1               float64 `json:"x1"`
	Y1               float64 `json:"y1"`
	Unit             string  `json:"unit,omitempty"`
	CoordinateSystem string  `json:"coordinate_system,omitempty"`
	Source           string  `json:"source,omitempty"`
}

// TextBlock is one ordered text observation with optional source geometry.
type TextBlock struct {
	Page       int         `json:"page,omitempty"`
	Text       string      `json:"text"`
	Confidence float64     `json:"confidence,omitempty"`
	Bounds     BoundingBox `json:"bounds,omitempty"`
}

// TableCell identifies one possibly merged table cell.
type TableCell struct {
	Page       int           `json:"page,omitempty"`
	TableIndex int           `json:"table_index,omitempty"`
	StartRow   int           `json:"start_row,omitempty"`
	StartCol   int           `json:"start_col,omitempty"`
	EndRow     int           `json:"end_row,omitempty"`
	EndCol     int           `json:"end_col,omitempty"`
	Text       string        `json:"text,omitempty"`
	Bounds     []BoundingBox `json:"bounds,omitempty"`
}

// ArtifactRef is a display-safe reference to a Host-published artifact.
type ArtifactRef struct {
	ID        string `json:"id,omitempty"`
	Kind      string `json:"kind,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Path      string `json:"path,omitempty"`
}
