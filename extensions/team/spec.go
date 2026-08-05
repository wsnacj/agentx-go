package team

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"
)

const (
	maxSpecBytes             = 256 << 10
	maxIdentityBytes         = 128
	maxNameBytes             = 512
	maxDescriptionBytes      = 16 << 10
	maxResponsibilityBytes   = 16 << 10
	maxMembers               = 64
	maxDependenciesPerMember = 64
	maxTags                  = 64
)

var forbiddenHostFields = map[string]bool{
	"approval": true, "budget": true, "concurrency": true, "credential": true,
	"credentials": true, "model": true, "provider": true, "queue": true,
	"retry": true, "runtime": true, "sandbox": true, "scheduler": true,
	"session": true, "timeout": true, "timeout_ms": true,
}

func Parse(content []byte) (Spec, error) {
	if len(content) == 0 || len(content) > maxSpecBytes {
		return Spec{}, specError(ErrorCodeInvalidSpec, "spec size is outside bounds")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil {
		return Spec{}, specError(ErrorCodeInvalidSpec, err.Error())
	}
	for field := range fields {
		if forbiddenHostFields[strings.ToLower(strings.TrimSpace(field))] {
			return Spec{}, specError(ErrorCodeForbiddenField, field)
		}
	}
	var raw Spec
	if err := json.Unmarshal(content, &raw); err != nil {
		return Spec{}, specError(ErrorCodeInvalidSpec, err.Error())
	}
	return Normalize(raw)
}

func Normalize(raw Spec) (Spec, error) {
	id := normalizeID(raw.ID)
	name := strings.TrimSpace(raw.Name)
	schema := strings.ToLower(strings.TrimSpace(raw.SchemaVersion))
	if schema == "" {
		schema = SchemaVersionV1
	}
	if schema != SchemaVersionV1 {
		return Spec{}, specError(ErrorCodeUnsupportedSchema, schema)
	}
	description := strings.TrimSpace(raw.Description)
	coordinator := normalizeID(raw.Coordinator)
	if id == "" || name == "" || coordinator == "" || len(name) > maxNameBytes || len(description) > maxDescriptionBytes ||
		len(raw.Members) == 0 || len(raw.Members) > maxMembers || len(raw.Tags) > maxTags {
		return Spec{}, specError(ErrorCodeInvalidSpec, "required field or size bound failed")
	}
	members := make([]Member, 0, len(raw.Members))
	seen := map[string]bool{}
	for _, item := range raw.Members {
		memberID := normalizeID(item.ID)
		expertID := normalizeID(item.ExpertID)
		responsibility := strings.TrimSpace(item.Responsibility)
		if memberID == "" || expertID == "" || seen[memberID] || len(responsibility) > maxResponsibilityBytes || len(item.DependsOn) > maxDependenciesPerMember {
			return Spec{}, specError(ErrorCodeInvalidSpec, "member is invalid or duplicated")
		}
		seen[memberID] = true
		members = append(members, Member{ID: memberID, ExpertID: expertID, Responsibility: responsibility, DependsOn: append([]string(nil), item.DependsOn...)})
	}
	if !seen[coordinator] {
		return Spec{}, specError(ErrorCodeInvalidSpec, "coordinator is not a member")
	}
	for index := range members {
		dependencies, err := normalizeDependencies(members[index].ID, members[index].DependsOn, seen)
		if err != nil {
			return Spec{}, err
		}
		members[index].DependsOn = dependencies
	}
	sort.Slice(members, func(i, j int) bool { return members[i].ID < members[j].ID })
	tags, err := normalizeTags(raw.Tags)
	if err != nil {
		return Spec{}, err
	}
	normalized := Spec{ID: id, SchemaVersion: schema, Name: name, Version: strings.TrimSpace(raw.Version), Description: description, Coordinator: coordinator, Members: members, Tags: tags}
	if _, err := buildNormalizedPlan(normalized); err != nil {
		return Spec{}, err
	}
	return normalized, nil
}

// BuildPlan validates a Team and returns stable topological stages. Members in
// one stage are independent from each other; execution policy remains Host-owned.
func BuildPlan(raw Spec) (Plan, error) {
	normalized, err := Normalize(raw)
	if err != nil {
		return Plan{}, err
	}
	return buildNormalizedPlan(normalized)
}

func Project(sourceRef string, raw Spec) (capabilitycatalog.Asset, error) {
	normalized, err := Normalize(raw)
	if err != nil {
		return capabilitycatalog.Asset{}, err
	}
	return capabilitycatalog.Asset{
		Identity: capabilitycatalog.Identity{Kind: capabilitycatalog.KindTeam, ID: normalized.ID},
		Name:     normalized.Name, Description: normalized.Description, Version: normalized.Version,
		SourceRef: strings.TrimSpace(sourceRef), Tags: append([]string(nil), normalized.Tags...),
	}, nil
}

func buildNormalizedPlan(spec Spec) (Plan, error) {
	members := make(map[string]Member, len(spec.Members))
	indegree := make(map[string]int, len(spec.Members))
	dependents := make(map[string][]string, len(spec.Members))
	for _, member := range spec.Members {
		members[member.ID] = member
		indegree[member.ID] = len(member.DependsOn)
		for _, dependency := range member.DependsOn {
			dependents[dependency] = append(dependents[dependency], member.ID)
		}
	}
	ready := make([]string, 0)
	for id, degree := range indegree {
		if degree == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	plan := Plan{TeamID: spec.ID, Coordinator: spec.Coordinator}
	processed := 0
	for len(ready) > 0 {
		stageIDs := append([]string(nil), ready...)
		stage := Stage{Number: len(plan.Stages) + 1, Members: make([]PlannedMember, 0, len(stageIDs))}
		next := make([]string, 0)
		for _, id := range stageIDs {
			member := members[id]
			stage.Members = append(stage.Members, PlannedMember{ID: member.ID, ExpertID: member.ExpertID, Responsibility: member.Responsibility})
			processed++
			for _, dependent := range dependents[id] {
				indegree[dependent]--
				if indegree[dependent] == 0 {
					next = append(next, dependent)
				}
			}
		}
		plan.Stages = append(plan.Stages, stage)
		sort.Strings(next)
		ready = next
	}
	if processed != len(spec.Members) {
		return Plan{}, specError(ErrorCodeDependencyCycle, "cycle detected")
	}
	return plan, nil
}

func normalizeDependencies(memberID string, raw []string, members map[string]bool) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		id := normalizeID(item)
		if id == "" || id == memberID || !members[id] || seen[id] {
			return nil, specError(ErrorCodeInvalidSpec, "member dependency is invalid")
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

func normalizeTags(raw []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" || len(value) > maxIdentityBytes {
			return nil, specError(ErrorCodeInvalidSpec, "tag is invalid")
		}
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out, nil
}

func normalizeID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || len(value) > maxIdentityBytes {
		return ""
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return ""
		}
	}
	return value
}

func specError(code ErrorCode, detail string) error {
	return &Error{Code: code, Cause: fmt.Errorf("%s", detail)}
}
