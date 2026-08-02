package diff

import "encoding/json"

type OCRLine struct {
	Angle               int         `json:"angle"`
	Text                string      `json:"text"`
	Score               float64     `json:"score"`
	Direction           int         `json:"direction,omitempty"`
	Handwritten         int         `json:"handwritten,omitempty"`
	Position            []int       `json:"position,omitempty"`
	Type                string      `json:"type,omitempty"`
	CharCandidates      [][]string  `json:"char_candidates,omitempty"`
	CharCandidatesScore [][]float64 `json:"char_candidates_score,omitempty"`
	CharCenters         [][]int     `json:"char_centers,omitempty"`
	CharPositions       [][]int     `json:"char_positions,omitempty"`
	CharScores          []float64   `json:"char_scores,omitempty"`
}

type OCRPage struct {
	Angle  int       `json:"angle"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Lines  []OCRLine `json:"lines"`
}

type OCRResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  struct {
		Pages []OCRPage `json:"pages"`
	} `json:"result"`
}

type TableResponse struct {
	Code   int    `json:"code"`
	Msg    string `json:"message"`
	Result struct {
		Pages []TPage `json:"pages"`
	} `json:"result"`
}

type TPage struct {
	Angle  int     `json:"angle"`
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Tables []Table `json:"tables"`
}

type Table struct {
	Lines        []TLine     `json:"lines"`
	TableCells   []TableCell `json:"table_cells"`
	TableRows    int         `json:"table_rows"`
	TableCols    int         `json:"table_cols"`
	HeightOfRows []int       `json:"height_of_rows"`
	WidthOfCols  []int       `json:"width_of_cols"`
	Position     []int       `json:"position"`
}

type TableCell struct {
	StartRow int     `json:"start_row"`
	StartCol int     `json:"start_col"`
	EndRow   int     `json:"end_row"`
	EndCol   int     `json:"end_col"`
	Text     string  `json:"text"`
	Lines    []TLine `json:"lines"`
	Position []int   `json:"position"`
}

type TLine struct {
	Angle               int         `json:"angle"`
	Text                string      `json:"text"`
	Direction           int         `json:"direction"`
	Handwritten         int         `json:"handwritten"`
	Position            []int       `json:"position"`
	Score               float64     `json:"score"`
	Type                string      `json:"type"`
	CharCandidates      [][]string  `json:"char_candidates,omitempty"`
	CharCandidatesScore [][]float64 `json:"char_candidates_score,omitempty"`
	CharPositions       [][]int     `json:"char_positions,omitempty"`
}

func ParseOCRResponse(data []byte) (*OCRResponse, error) {
	var resp OCRResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func ParseTableResponse(data []byte) (*TableResponse, error) {
	var resp TableResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
