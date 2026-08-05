package team

const SchemaVersionV1 = "v1"

type Member struct {
	ID             string   `json:"id"`
	ExpertID       string   `json:"expert_id"`
	Responsibility string   `json:"responsibility,omitempty"`
	DependsOn      []string `json:"depends_on,omitempty"`
}

type Spec struct {
	ID            string   `json:"id"`
	SchemaVersion string   `json:"schema_version"`
	Name          string   `json:"name"`
	Version       string   `json:"version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Coordinator   string   `json:"coordinator"`
	Members       []Member `json:"members"`
	Tags          []string `json:"tags,omitempty"`
}

type PlannedMember struct {
	ID             string `json:"id"`
	ExpertID       string `json:"expert_id"`
	Responsibility string `json:"responsibility,omitempty"`
}

type Stage struct {
	Number  int             `json:"number"`
	Members []PlannedMember `json:"members"`
}

// Plan is a deterministic topological view. It carries no worker, queue,
// model, budget, retry or execution state.
type Plan struct {
	TeamID      string  `json:"team_id"`
	Coordinator string  `json:"coordinator"`
	Stages      []Stage `json:"stages"`
}
