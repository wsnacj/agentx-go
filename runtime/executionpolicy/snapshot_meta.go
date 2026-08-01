package executionpolicy

import (
	"encoding/json"
	"strings"
)

const sessionMetaExecutionContractKey = "agentx_execution_contract"

func LoadSnapshotMetaJSON(metaJSON string) (Snapshot, bool) {
	trimmed := strings.TrimSpace(metaJSON)
	if trimmed == "" {
		return Snapshot{}, false
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return Snapshot{}, false
	}
	raw, ok := payload[sessionMetaExecutionContractKey]
	if !ok || raw == nil {
		return Snapshot{}, false
	}
	blob, err := json.Marshal(raw)
	if err != nil {
		return Snapshot{}, false
	}
	var out Snapshot
	if err := json.Unmarshal(blob, &out); err != nil {
		return Snapshot{}, false
	}
	return out, true
}

func MergeSnapshotMetaJSON(metaJSON string, snapshot Snapshot) (string, error) {
	payload := map[string]any{}
	trimmed := strings.TrimSpace(metaJSON)
	if trimmed != "" {
		if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
			return "", err
		}
	}
	payload[sessionMetaExecutionContractKey] = snapshot
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}
