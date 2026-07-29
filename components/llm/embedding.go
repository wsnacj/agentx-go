package llm

// EmbeddingRequest holds either text or multimodal inputs.
type EmbeddingRequest struct {
	Model     string
	Inputs    []string
	Images    []string
	Path      string
	BatchSize int
	Encoding  string
	Options   map[string]any
}

// SparseEntry represents a sparse vector entry.
type SparseEntry struct {
	Index int     `json:"index"`
	Value float32 `json:"value"`
}

// EmbeddingResponse contains one or more vectors.
type EmbeddingResponse struct {
	Vectors       [][]float32
	SparseVectors [][]SparseEntry
	Raw           []byte
}
