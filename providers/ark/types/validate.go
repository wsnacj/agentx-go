package types

import (
	"errors"
	"fmt"
)

const (
	ToolTypeFunction        = "function"
	ToolTypeWebSearch       = "web_search"
	ToolTypeImageProcess    = "image_process"
	ToolTypeKnowledgeSearch = "knowledge_search"
	ToolTypeMCP             = "mcp"
	ToolTypeDoubaoApp       = "doubao_app"

	ToggleEnabled  = "enabled"
	ToggleDisabled = "disabled"
)

// ValidateTools validates a tool list and returns a combined error.
func ValidateTools(tools []Tool) error {
	if len(tools) == 0 {
		return nil
	}
	var errs []error
	for i, tool := range tools {
		if err := ValidateTool(tool); err != nil {
			errs = append(errs, fmt.Errorf("tool[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// ValidateTool checks required fields based on tool type.
func ValidateTool(tool Tool) error {
	if tool.Type == "" {
		return fmt.Errorf("tool type is required")
	}
	var errs []error
	switch tool.Type {
	case ToolTypeFunction:
		if tool.Name == "" {
			errs = append(errs, fmt.Errorf("function name is required"))
		}
		if len(tool.Parameters) == 0 {
			errs = append(errs, fmt.Errorf("function parameters are required"))
		}
	case ToolTypeKnowledgeSearch:
		if tool.KnowledgeSearchID == "" {
			errs = append(errs, fmt.Errorf("knowledge_search_id is required"))
		}
		if tool.Limit != nil && (*tool.Limit < 1 || *tool.Limit > 200) {
			errs = append(errs, fmt.Errorf("limit must be in [1,200]"))
		}
		if tool.MaxKeyword != nil && (*tool.MaxKeyword < 1 || *tool.MaxKeyword > 50) {
			errs = append(errs, fmt.Errorf("max_keyword must be in [1,50]"))
		}
		if tool.DenseWeight != nil && (*tool.DenseWeight < 0.2 || *tool.DenseWeight > 1) {
			errs = append(errs, fmt.Errorf("dense_weight must be in [0.2,1]"))
		}
	case ToolTypeWebSearch:
		if tool.Limit != nil && (*tool.Limit < 1 || *tool.Limit > 50) {
			errs = append(errs, fmt.Errorf("limit must be in [1,50]"))
		}
		if tool.MaxKeyword != nil && (*tool.MaxKeyword < 1 || *tool.MaxKeyword > 50) {
			errs = append(errs, fmt.Errorf("max_keyword must be in [1,50]"))
		}
	case ToolTypeMCP:
		if tool.ServerLabel == "" {
			errs = append(errs, fmt.Errorf("server_label is required"))
		}
		if tool.ServerURL == "" {
			errs = append(errs, fmt.Errorf("server_url is required"))
		}
	case ToolTypeImageProcess:
		if err := validateToggle("point", tool.Point); err != nil {
			errs = append(errs, err)
		}
		if err := validateToggle("grounding", tool.Grounding); err != nil {
			errs = append(errs, err)
		}
		if err := validateToggle("zoom", tool.Zoom); err != nil {
			errs = append(errs, err)
		}
		if err := validateToggle("rotate", tool.Rotate); err != nil {
			errs = append(errs, err)
		}
	case ToolTypeDoubaoApp:
		if err := validateFeatureToggle("chat", tool.Feature, func(f *DoubaoAppFeature) *FeatureToggle { return f.Chat }); err != nil {
			errs = append(errs, err)
		}
		if err := validateFeatureToggle("deep_chat", tool.Feature, func(f *DoubaoAppFeature) *FeatureToggle { return f.DeepChat }); err != nil {
			errs = append(errs, err)
		}
		if err := validateFeatureToggle("ai_search", tool.Feature, func(f *DoubaoAppFeature) *FeatureToggle { return f.AISearch }); err != nil {
			errs = append(errs, err)
		}
		if err := validateFeatureToggle("reasoning_search", tool.Feature, func(f *DoubaoAppFeature) *FeatureToggle { return f.ReasoningSearch }); err != nil {
			errs = append(errs, err)
		}
	default:
		return nil
	}
	if err := validateUserLocation(tool.UserLocation); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func validateToggle(name string, cfg *ToggleConfig) error {
	if cfg == nil || cfg.Type == "" {
		return nil
	}
	if cfg.Type != ToggleEnabled && cfg.Type != ToggleDisabled {
		return fmt.Errorf("%s type must be enabled or disabled", name)
	}
	return nil
}

func validateFeatureToggle(name string, feature *DoubaoAppFeature, getter func(*DoubaoAppFeature) *FeatureToggle) error {
	if feature == nil {
		return nil
	}
	cfg := getter(feature)
	if cfg == nil || cfg.Type == "" {
		return nil
	}
	if cfg.Type != ToggleEnabled && cfg.Type != ToggleDisabled {
		return fmt.Errorf("%s type must be enabled or disabled", name)
	}
	return nil
}

func validateUserLocation(location *UserLocation) error {
	if location == nil || location.Type == "" {
		return nil
	}
	if location.Country == "" && location.Region == "" && location.City == "" {
		return fmt.Errorf("user_location requires country, region, or city")
	}
	return nil
}
