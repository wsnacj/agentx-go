package types

import "encoding/json"

// Tool describes a Responses API tool definition.
type Tool struct {
	Type              string            `json:"type"`
	Name              string            `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	Parameters        map[string]any    `json:"parameters,omitempty"`
	Strict            *bool             `json:"strict,omitempty"`
	Feature           *DoubaoAppFeature `json:"feature,omitempty"`
	Sources           []string          `json:"sources,omitempty"`
	Limit             *int              `json:"limit,omitempty"`
	MaxKeyword        *int              `json:"max_keyword,omitempty"`
	UserLocation      *UserLocation     `json:"user_location,omitempty"`
	KnowledgeSearchID string            `json:"knowledge_search_id,omitempty"`
	DocFilters        map[string]any    `json:"doc_filters,omitempty"`
	DenseWeight       *float64          `json:"dense_weight,omitempty"`
	RankingOptions    *RankingOptions   `json:"ranking_options,omitempty"`
	Point             *ToggleConfig     `json:"point,omitempty"`
	Grounding         *ToggleConfig     `json:"grounding,omitempty"`
	Zoom              *ToggleConfig     `json:"zoom,omitempty"`
	Rotate            *ToggleConfig     `json:"rotate,omitempty"`
	ServerLabel       string            `json:"server_label,omitempty"`
	ServerURL         string            `json:"server_url,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	RequireApproval   any               `json:"require_approval,omitempty"`
	AllowedTools      any               `json:"allowed_tools,omitempty"`
	Extra             map[string]any    `json:"-"`
}

// MarshalJSON allows custom tool extensions.
func (t Tool) MarshalJSON() ([]byte, error) {
	payload := map[string]any{}
	data, err := json.Marshal(struct {
		Type              string            `json:"type"`
		Name              string            `json:"name,omitempty"`
		Description       string            `json:"description,omitempty"`
		Parameters        map[string]any    `json:"parameters,omitempty"`
		Strict            *bool             `json:"strict,omitempty"`
		Feature           *DoubaoAppFeature `json:"feature,omitempty"`
		Sources           []string          `json:"sources,omitempty"`
		Limit             *int              `json:"limit,omitempty"`
		MaxKeyword        *int              `json:"max_keyword,omitempty"`
		UserLocation      *UserLocation     `json:"user_location,omitempty"`
		KnowledgeSearchID string            `json:"knowledge_search_id,omitempty"`
		DocFilters        map[string]any    `json:"doc_filters,omitempty"`
		DenseWeight       *float64          `json:"dense_weight,omitempty"`
		RankingOptions    *RankingOptions   `json:"ranking_options,omitempty"`
		Point             *ToggleConfig     `json:"point,omitempty"`
		Grounding         *ToggleConfig     `json:"grounding,omitempty"`
		Zoom              *ToggleConfig     `json:"zoom,omitempty"`
		Rotate            *ToggleConfig     `json:"rotate,omitempty"`
		ServerLabel       string            `json:"server_label,omitempty"`
		ServerURL         string            `json:"server_url,omitempty"`
		Headers           map[string]string `json:"headers,omitempty"`
		RequireApproval   any               `json:"require_approval,omitempty"`
		AllowedTools      any               `json:"allowed_tools,omitempty"`
	}{
		Type:              t.Type,
		Name:              t.Name,
		Description:       t.Description,
		Parameters:        t.Parameters,
		Strict:            t.Strict,
		Feature:           t.Feature,
		Sources:           t.Sources,
		Limit:             t.Limit,
		MaxKeyword:        t.MaxKeyword,
		UserLocation:      t.UserLocation,
		KnowledgeSearchID: t.KnowledgeSearchID,
		DocFilters:        t.DocFilters,
		DenseWeight:       t.DenseWeight,
		RankingOptions:    t.RankingOptions,
		Point:             t.Point,
		Grounding:         t.Grounding,
		Zoom:              t.Zoom,
		Rotate:            t.Rotate,
		ServerLabel:       t.ServerLabel,
		ServerURL:         t.ServerURL,
		Headers:           t.Headers,
		RequireApproval:   t.RequireApproval,
		AllowedTools:      t.AllowedTools,
	})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	for k, v := range t.Extra {
		payload[k] = v
	}
	return json.Marshal(payload)
}

// DoubaoAppFeature configures doubao_app tool behavior.
type DoubaoAppFeature struct {
	Chat            *FeatureToggle `json:"chat,omitempty"`
	DeepChat        *FeatureToggle `json:"deep_chat,omitempty"`
	AISearch        *FeatureToggle `json:"ai_search,omitempty"`
	ReasoningSearch *FeatureToggle `json:"reasoning_search,omitempty"`
}

// FeatureToggle toggles a feature with optional role description.
type FeatureToggle struct {
	Type            string `json:"type,omitempty"`
	RoleDescription string `json:"role_description,omitempty"`
}

// UserLocation describes user location metadata.
type UserLocation struct {
	Type    string `json:"type,omitempty"`
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
}

// ToggleConfig toggles a tool capability.
type ToggleConfig struct {
	Type string `json:"type,omitempty"`
}

// RankingOptions configures knowledge search ranking.
type RankingOptions struct {
	RerankSwitch        *bool  `json:"rerank_switch,omitempty"`
	RetrieveCount       *int   `json:"retrieve_count,omitempty"`
	GetAttachmentLink   *bool  `json:"get_attachment_link,omitempty"`
	ChunkDiffusionCount *int   `json:"chunk_diffusion_count,omitempty"`
	ChunkGroup          *bool  `json:"chunk_group,omitempty"`
	RerankModel         string `json:"rerank_model,omitempty"`
	RerankOnlyChunk     *bool  `json:"rerank_only_chunk,omitempty"`
}
