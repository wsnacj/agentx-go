package types

import (
	"encoding/json"
	"fmt"
)

// ImageInput supports single-image or multi-image inputs.
type ImageInput struct {
	Single *string
	Multi  []string
}

// NewImageInput creates a single-image input.
func NewImageInput(image string) ImageInput {
	return ImageInput{Single: &image}
}

// NewImageInputList creates a multi-image input.
func NewImageInputList(images ...string) ImageInput {
	if len(images) == 0 {
		return ImageInput{}
	}
	copied := make([]string, len(images))
	copy(copied, images)
	return ImageInput{Multi: copied}
}

func (i ImageInput) MarshalJSON() ([]byte, error) {
	if i.Single != nil {
		return json.Marshal(*i.Single)
	}
	if i.Multi != nil {
		return json.Marshal(i.Multi)
	}
	return []byte("null"), nil
}

func (i *ImageInput) UnmarshalJSON(data []byte) error {
	if i == nil {
		return fmt.Errorf("ark types: image input is nil")
	}
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		i.Single = &one
		i.Multi = nil
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		i.Multi = many
		i.Single = nil
		return nil
	}
	return fmt.Errorf("ark types: unsupported image input")
}

// ImageGenerationTool defines a tool for image generation.
type ImageGenerationTool struct {
	Type string `json:"type"`
}

// SequentialImageGenerationOptions configures the image sequence output.
type SequentialImageGenerationOptions struct {
	MaxImages int `json:"max_images,omitempty"`
}

// OptimizePromptOptions controls prompt optimization behavior.
type OptimizePromptOptions struct {
	Mode string `json:"mode,omitempty"`
}

// ImageGenerationRequest mirrors Ark Images API request fields.
type ImageGenerationRequest struct {
	Model                            string                            `json:"model"`
	Prompt                           string                            `json:"prompt"`
	Image                            ImageInput                        `json:"image,omitempty"`
	Size                             string                            `json:"size,omitempty"`
	Seed                             *int                              `json:"seed,omitempty"`
	SequentialImageGeneration        string                            `json:"sequential_image_generation,omitempty"`
	SequentialImageGenerationOptions *SequentialImageGenerationOptions `json:"sequential_image_generation_options,omitempty"`
	Tools                            []ImageGenerationTool             `json:"tools,omitempty"`
	Stream                           *bool                             `json:"stream,omitempty"`
	GuidanceScale                    *float64                          `json:"guidance_scale,omitempty"`
	OutputFormat                     string                            `json:"output_format,omitempty"`
	ResponseFormat                   string                            `json:"response_format,omitempty"`
	Watermark                        *bool                             `json:"watermark,omitempty"`
	OptimizePromptOptions            *OptimizePromptOptions            `json:"optimize_prompt_options,omitempty"`
	Extra                            map[string]any                    `json:"-"`
	ExtraHeaders                     map[string]string                 `json:"-"`
}

// MergeExtra merges custom fields into the request payload.
func (r ImageGenerationRequest) MarshalJSON() ([]byte, error) {
	payload := map[string]any{}
	data, err := json.Marshal(struct {
		Model                            string                            `json:"model"`
		Prompt                           string                            `json:"prompt"`
		Size                             string                            `json:"size,omitempty"`
		Seed                             *int                              `json:"seed,omitempty"`
		SequentialImageGeneration        string                            `json:"sequential_image_generation,omitempty"`
		SequentialImageGenerationOptions *SequentialImageGenerationOptions `json:"sequential_image_generation_options,omitempty"`
		Tools                            []ImageGenerationTool             `json:"tools,omitempty"`
		Stream                           *bool                             `json:"stream,omitempty"`
		GuidanceScale                    *float64                          `json:"guidance_scale,omitempty"`
		OutputFormat                     string                            `json:"output_format,omitempty"`
		ResponseFormat                   string                            `json:"response_format,omitempty"`
		Watermark                        *bool                             `json:"watermark,omitempty"`
		OptimizePromptOptions            *OptimizePromptOptions            `json:"optimize_prompt_options,omitempty"`
	}{
		Model:                            r.Model,
		Prompt:                           r.Prompt,
		Size:                             r.Size,
		Seed:                             r.Seed,
		SequentialImageGeneration:        r.SequentialImageGeneration,
		SequentialImageGenerationOptions: r.SequentialImageGenerationOptions,
		Tools:                            r.Tools,
		Stream:                           r.Stream,
		GuidanceScale:                    r.GuidanceScale,
		OutputFormat:                     r.OutputFormat,
		ResponseFormat:                   r.ResponseFormat,
		Watermark:                        r.Watermark,
		OptimizePromptOptions:            r.OptimizePromptOptions,
	})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if hasImageInput(r.Image) {
		payload["image"] = r.Image
	}
	for k, v := range r.Extra {
		if k == "" {
			continue
		}
		payload[k] = v
	}
	return json.Marshal(payload)
}

func hasImageInput(input ImageInput) bool {
	if input.Single != nil {
		return true
	}
	return len(input.Multi) > 0
}
