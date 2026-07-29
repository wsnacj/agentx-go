package toolerrors

import (
	"errors"
	"strings"
)

const (
	// ToolArgumentErrorCodeInvalidJSON marks syntactically invalid JSON arguments.
	ToolArgumentErrorCodeInvalidJSON = "invalid_json"
	// ToolArgumentErrorCodeInvalidArgumentObject marks valid JSON with a non-object top level.
	ToolArgumentErrorCodeInvalidArgumentObject = "invalid_argument_object"
	// ToolArgumentErrorCodeInvalidArgument marks invalid field values in otherwise decoded arguments.
	ToolArgumentErrorCodeInvalidArgument = "invalid_argument"
	// ToolArgumentErrorCodeMissingRequiredArgument marks absent required fields.
	ToolArgumentErrorCodeMissingRequiredArgument = "missing_required_argument"

	// ToolArgumentRepairReturnValidJSONObject asks the model to resend a valid JSON object.
	ToolArgumentRepairReturnValidJSONObject = "return_valid_json_object"
	// ToolArgumentRepairProvideRequiredField asks the model to provide missing required fields.
	ToolArgumentRepairProvideRequiredField = "provide_required_field"
	// ToolArgumentRepairFixInvalidField asks the model to correct invalid fields.
	ToolArgumentRepairFixInvalidField = "fix_invalid_field"
	// ToolArgumentRepairUseAliasURL promotes a tool-declared URL alias to url.
	ToolArgumentRepairUseAliasURL = "use_alias_url"
)

// ToolArgumentRepair describes a deterministic, tool-declared repair kind.
type ToolArgumentRepair struct {
	Kind string `json:"kind,omitempty"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// ToolArgumentErrorOptions controls a structured invalid-args error.
type ToolArgumentErrorOptions struct {
	Code             string
	Detail           string
	Repairable       bool
	SafeAutorepair   bool
	MissingFields    []string
	InvalidFields    []string
	DisallowedFields []string
	AllowedRepairs   []ToolArgumentRepair
	Cause            error
}

// ToolArgumentError represents a structured invalid tool-argument failure.
type ToolArgumentError struct {
	Tool             string
	Code             string
	Detail           string
	Repairable       bool
	SafeAutorepair   bool
	MissingFields    []string
	InvalidFields    []string
	DisallowedFields []string
	AllowedRepairs   []ToolArgumentRepair
	Cause            error
}

func (e *ToolArgumentError) Error() string {
	if e == nil {
		return "invalid arguments"
	}
	if detail := strings.TrimSpace(e.Detail); detail != "" {
		return detail
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "invalid arguments"
}

func (e *ToolArgumentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func AsToolArgumentError(err error) (*ToolArgumentError, bool) {
	var target *ToolArgumentError
	if !errors.As(err, &target) {
		return nil, false
	}
	return target, true
}

func NewToolArgumentError(tool string, opts ToolArgumentErrorOptions) error {
	return &ToolArgumentError{
		Tool:             strings.TrimSpace(tool),
		Code:             strings.TrimSpace(opts.Code),
		Detail:           strings.TrimSpace(opts.Detail),
		Repairable:       opts.Repairable,
		SafeAutorepair:   opts.SafeAutorepair,
		MissingFields:    append([]string(nil), opts.MissingFields...),
		InvalidFields:    append([]string(nil), opts.InvalidFields...),
		DisallowedFields: append([]string(nil), opts.DisallowedFields...),
		AllowedRepairs:   append([]ToolArgumentRepair(nil), opts.AllowedRepairs...),
		Cause:            opts.Cause,
	}
}

// NewInvalidJSONToolArgumentError wraps shared JSON decode failures as structured argument errors.
func NewInvalidJSONToolArgumentError(tool string, cause error) error {
	detail := "decode tool args: invalid json"
	if cause != nil {
		detail = cause.Error()
	}
	code := ToolArgumentErrorCodeInvalidJSON
	if strings.Contains(strings.ToLower(detail), "top-level json object is required") {
		code = ToolArgumentErrorCodeInvalidArgumentObject
	}
	return NewToolArgumentError(tool, ToolArgumentErrorOptions{
		Code:           code,
		Detail:         detail,
		Repairable:     true,
		SafeAutorepair: false,
		AllowedRepairs: []ToolArgumentRepair{{Kind: ToolArgumentRepairReturnValidJSONObject}},
		Cause:          cause,
	})
}

// NewInvalidToolArgumentError reports invalid decoded argument field values.
func NewInvalidToolArgumentError(tool string, fields []string, detail string) error {
	cleanFields := cleanArgumentFieldNames(fields)
	if strings.TrimSpace(detail) == "" {
		if len(cleanFields) == 1 {
			detail = strings.TrimSpace(tool) + ": " + cleanFields[0] + " is invalid"
		} else {
			detail = strings.TrimSpace(tool) + ": invalid arguments"
		}
	}
	return NewToolArgumentError(tool, ToolArgumentErrorOptions{
		Code:           ToolArgumentErrorCodeInvalidArgument,
		Detail:         detail,
		Repairable:     true,
		SafeAutorepair: false,
		InvalidFields:  cleanFields,
		AllowedRepairs: []ToolArgumentRepair{{Kind: ToolArgumentRepairFixInvalidField, To: strings.Join(cleanFields, ",")}},
	})
}

// NewMissingRequiredToolArgumentError reports absent required argument fields.
func NewMissingRequiredToolArgumentError(tool string, fields []string, detail string) error {
	cleanFields := cleanArgumentFieldNames(fields)
	if strings.TrimSpace(detail) == "" {
		if len(cleanFields) == 1 {
			detail = strings.TrimSpace(tool) + ": " + cleanFields[0] + " is required"
		} else {
			detail = strings.TrimSpace(tool) + ": required arguments are missing"
		}
	}
	return NewToolArgumentError(tool, ToolArgumentErrorOptions{
		Code:           ToolArgumentErrorCodeMissingRequiredArgument,
		Detail:         detail,
		Repairable:     true,
		SafeAutorepair: false,
		MissingFields:  cleanFields,
		AllowedRepairs: []ToolArgumentRepair{{Kind: ToolArgumentRepairProvideRequiredField, To: strings.Join(cleanFields, ",")}},
	})
}

func cleanArgumentFieldNames(fields []string) []string {
	if len(fields) == 0 {
		return nil
	}
	out := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}
