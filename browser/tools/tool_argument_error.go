package tools

import agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"

const (
	ToolArgumentErrorCodeInvalidJSON             = agentxtoolerrors.ToolArgumentErrorCodeInvalidJSON
	ToolArgumentErrorCodeInvalidArgumentObject   = agentxtoolerrors.ToolArgumentErrorCodeInvalidArgumentObject
	ToolArgumentErrorCodeInvalidArgument         = agentxtoolerrors.ToolArgumentErrorCodeInvalidArgument
	ToolArgumentErrorCodeMissingRequiredArgument = agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument
	ToolArgumentRepairReturnValidJSONObject      = agentxtoolerrors.ToolArgumentRepairReturnValidJSONObject
	ToolArgumentRepairProvideRequiredField       = agentxtoolerrors.ToolArgumentRepairProvideRequiredField
	ToolArgumentRepairFixInvalidField            = agentxtoolerrors.ToolArgumentRepairFixInvalidField
	ToolArgumentRepairUseAliasURL                = agentxtoolerrors.ToolArgumentRepairUseAliasURL
)

type ToolArgumentRepair = agentxtoolerrors.ToolArgumentRepair
type ToolArgumentErrorOptions = agentxtoolerrors.ToolArgumentErrorOptions
type ToolArgumentError = agentxtoolerrors.ToolArgumentError

func AsToolArgumentError(err error) (*ToolArgumentError, bool) {
	return agentxtoolerrors.AsToolArgumentError(err)
}

func NewToolArgumentError(tool string, opts ToolArgumentErrorOptions) error {
	return agentxtoolerrors.NewToolArgumentError(tool, opts)
}

func NewInvalidJSONToolArgumentError(tool string, cause error) error {
	return agentxtoolerrors.NewInvalidJSONToolArgumentError(tool, cause)
}

func NewInvalidToolArgumentError(tool string, fields []string, detail string) error {
	return agentxtoolerrors.NewInvalidToolArgumentError(tool, fields, detail)
}

func NewMissingRequiredToolArgumentError(tool string, fields []string, detail string) error {
	return agentxtoolerrors.NewMissingRequiredToolArgumentError(tool, fields, detail)
}

func rejectHostOwnedToolArguments(tool string, params map[string]any, fields ...string) error {
	present := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := params[field]; ok {
			present = append(present, field)
		}
	}
	if len(present) == 0 {
		return nil
	}
	return NewInvalidToolArgumentError(
		tool,
		present,
		tool+": workspace, memory, and artifact roots are host-owned and cannot be overridden by tool arguments",
	)
}
