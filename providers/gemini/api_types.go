package gemini

import (
	"encoding/json"
	"fmt"
)

// Role constants for content items.
const (
	RoleUser  = "user"
	RoleModel = "model"
)

// Content represents a single conversation turn.
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

// Part represents a multi-modal part. Only one of the data fields should be set.
type Part struct {
	Text             string            `json:"text,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	InlineData       *InlineData       `json:"inlineData,omitempty"`
	FileData         *FileData         `json:"fileData,omitempty"`
	VideoMetadata    *VideoMetadata    `json:"videoMetadata,omitempty"`
	MediaResolution  *MediaResolution  `json:"mediaResolution,omitempty"`
}

// FunctionCall is a Gemini-native function invocation emitted by a model.
type FunctionCall struct {
	Name string         `json:"name,omitempty"`
	Args map[string]any `json:"args,omitempty"`
}

// FunctionResponse is a Gemini-native function result supplied by the Host.
type FunctionResponse struct {
	Name     string         `json:"name,omitempty"`
	Response map[string]any `json:"response,omitempty"`
}

// MediaResolution controls media token budget at part level.
type MediaResolution struct {
	Level string `json:"level,omitempty"`
}

// InlineData holds base64-encoded bytes.
type InlineData struct {
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

// FileData references a file uploaded to the Files API.
type FileData struct {
	MimeType string `json:"mimeType,omitempty"`
	FileURI  string `json:"fileUri,omitempty"`
}

// VideoMetadata controls video sampling.
type VideoMetadata struct {
	StartOffset *Duration `json:"startOffset,omitempty"`
	EndOffset   *Duration `json:"endOffset,omitempty"`
	FPS         *float32  `json:"fps,omitempty"`
}

// Duration represents a protobuf Duration.
type Duration struct {
	Seconds int64 `json:"seconds,omitempty"`
	Nanos   int32 `json:"nanos,omitempty"`
}

// MarshalJSON serializes protobuf Duration as "<seconds>s" for Gemini API compatibility.
func (d Duration) MarshalJSON() ([]byte, error) {
	if d.Nanos == 0 {
		return json.Marshal(fmt.Sprintf("%ds", d.Seconds))
	}
	if d.Nanos < 0 {
		return json.Marshal(fmt.Sprintf("%d.%09ds", d.Seconds, -d.Nanos))
	}
	return json.Marshal(fmt.Sprintf("%d.%09ds", d.Seconds, d.Nanos))
}

// GenerationConfig tunes model output.
type GenerationConfig struct {
	CandidateCount     *int            `json:"candidateCount,omitempty"`
	StopSequences      []string        `json:"stopSequences,omitempty"`
	MaxOutputTokens    *int            `json:"maxOutputTokens,omitempty"`
	Temperature        *float32        `json:"temperature,omitempty"`
	TopP               *float32        `json:"topP,omitempty"`
	TopK               *int            `json:"topK,omitempty"`
	PresencePenalty    *float32        `json:"presencePenalty,omitempty"`
	FrequencyPenalty   *float32        `json:"frequencyPenalty,omitempty"`
	ResponseMimeType   string          `json:"responseMimeType,omitempty"`
	ResponseSchema     any             `json:"responseSchema,omitempty"`
	ResponseJSONSchema any             `json:"responseJsonSchema,omitempty"`
	MediaResolution    string          `json:"mediaResolution,omitempty"`
	ThinkingConfig     *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

// ThinkingConfig controls Gemini thinking depth.
type ThinkingConfig struct {
	ThinkingLevel   string `json:"thinkingLevel,omitempty"`
	IncludeThoughts *bool  `json:"includeThoughts,omitempty"`
}

// SafetySetting configures per-category safety thresholds.
type SafetySetting struct {
	Category  string `json:"category,omitempty"`
	Threshold string `json:"threshold,omitempty"`
}

// Tool defines tool configuration in the request.
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
	GoogleSearch         *GoogleSearchTool     `json:"googleSearch,omitempty"`
	URLContext           *URLContextTool       `json:"urlContext,omitempty"`
	CodeExecution        *CodeExecutionTool    `json:"codeExecution,omitempty"`
}

// GoogleSearchTool enables grounded web search.
type GoogleSearchTool struct{}

// URLContextTool enables URL context retrieval.
type URLContextTool struct{}

// CodeExecutionTool enables sandboxed code execution.
type CodeExecutionTool struct{}

// ToolConfig controls tool usage behavior.
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// FunctionDeclaration describes a callable function.
type FunctionDeclaration struct {
	Name                 string         `json:"name,omitempty"`
	Description          string         `json:"description,omitempty"`
	Parameters           map[string]any `json:"parameters,omitempty"`
	ParametersJSONSchema any            `json:"parametersJsonSchema,omitempty"`
	ResponseJSONSchema   any            `json:"responseJsonSchema,omitempty"`
}

// FunctionCallingConfig configures function calling.
type FunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

// GenerateContentRequest represents a generateContent call.
type GenerateContentRequest struct {
	Contents          []Content         `json:"contents,omitempty"`
	SystemInstruction *Content          `json:"systemInstruction,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings    []SafetySetting   `json:"safetySettings,omitempty"`
	Tools             []Tool            `json:"tools,omitempty"`
	ToolConfig        *ToolConfig       `json:"toolConfig,omitempty"`
	CachedContent     string            `json:"cachedContent,omitempty"`
}

// GenerateContentResponse mirrors the API response.
type GenerateContentResponse struct {
	Candidates     []Candidate     `json:"candidates,omitempty"`
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
}

// Candidate represents a single model output.
type Candidate struct {
	Content       Content        `json:"content,omitempty"`
	FinishReason  string         `json:"finishReason,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// SafetyRating represents a safety assessment.
type SafetyRating struct {
	Category    string `json:"category,omitempty"`
	Probability string `json:"probability,omitempty"`
}

// PromptFeedback represents feedback about the prompt.
type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// UsageMetadata contains token counts.
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
}

// EmbedContentRequest represents a single embedding request.
type EmbedContentRequest struct {
	Content              Content `json:"content,omitempty"`
	OutputDimensionality *int    `json:"outputDimensionality,omitempty"`
	TaskType             string  `json:"taskType,omitempty"`
	Title                string  `json:"title,omitempty"`
}

// EmbedContentResponse represents embedding output.
type EmbedContentResponse struct {
	Embedding Embedding `json:"embedding,omitempty"`
}

// Embedding is a vector output.
type Embedding struct {
	Values []float32 `json:"values,omitempty"`
}

// BatchEmbedContentsRequest represents a batch embedding request.
type BatchEmbedContentsRequest struct {
	Requests []EmbedContentRequest `json:"requests,omitempty"`
}

// BatchEmbedContentsResponse represents batch embedding output.
type BatchEmbedContentsResponse struct {
	Embeddings []Embedding `json:"embeddings,omitempty"`
}

// File represents an uploaded file.
type File struct {
	Name           string        `json:"name,omitempty"`
	DisplayName    string        `json:"displayName,omitempty"`
	MimeType       string        `json:"mimeType,omitempty"`
	SizeBytes      string        `json:"sizeBytes,omitempty"`
	CreateTime     string        `json:"createTime,omitempty"`
	UpdateTime     string        `json:"updateTime,omitempty"`
	ExpirationTime string        `json:"expirationTime,omitempty"`
	SHA256Hash     string        `json:"sha256Hash,omitempty"`
	URI            string        `json:"uri,omitempty"`
	DownloadURI    string        `json:"downloadUri,omitempty"`
	State          string        `json:"state,omitempty"`
	Source         string        `json:"source,omitempty"`
	Error          *Status       `json:"error,omitempty"`
	Metadata       *FileMetadata `json:"metadata,omitempty"`
}

// FileMetadata holds metadata for specific file types.
type FileMetadata struct {
	VideoMetadata *VideoFileMetadata `json:"videoMetadata,omitempty"`
}

// VideoFileMetadata holds metadata for video files.
type VideoFileMetadata struct {
	DurationSeconds string `json:"durationSeconds,omitempty"`
}

// Status represents error status for file processing.
type Status struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// FileResponse wraps a file object.
type FileResponse struct {
	File *File `json:"file,omitempty"`
}

// Helper constructors.
func NewContent(parts ...Part) Content {
	return Content{Role: RoleUser, Parts: parts}
}

func NewContentWithRole(role string, parts ...Part) Content {
	return Content{Role: role, Parts: parts}
}

func NewTextPart(text string) Part {
	return Part{Text: text}
}

func NewInlineDataPart(mimeType string, data string) Part {
	return Part{InlineData: &InlineData{MimeType: mimeType, Data: data}}
}

func NewFileDataPart(fileURI string, mimeType string) Part {
	return Part{FileData: &FileData{FileURI: fileURI, MimeType: mimeType}}
}

func NewVideoPartFromURI(fileURI string, mimeType string, fps *float32) Part {
	part := NewFileDataPart(fileURI, mimeType)
	if fps != nil {
		part.VideoMetadata = &VideoMetadata{FPS: fps}
	}
	return part
}

func NewSystemInstruction(text string) *Content {
	if text == "" {
		return nil
	}
	return &Content{Parts: []Part{{Text: text}}}
}

func ResponseText(resp *GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	parts := resp.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return ""
	}
	out := ""
	for _, p := range parts {
		if p.Text == "" {
			continue
		}
		out += p.Text
	}
	return out
}
