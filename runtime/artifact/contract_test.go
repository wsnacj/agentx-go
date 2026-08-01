package artifact

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestArtifactJSONContracts(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  string
	}{
		{name: "record zero", value: Record{}, want: `{"artifact_id":"","created_at":0}`},
		{name: "link zero", value: Link{}, want: `{"source_artifact_id":"","target_artifact_id":"","relation":"","created_at":0}`},
		{name: "link filter zero", value: LinkFilter{}, want: `{}`},
		{name: "blob input hides data", value: BlobPutInput{Data: []byte("private")}, want: `{}`},
		{name: "blob ref zero", value: BlobRef{}, want: `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}
			if string(encoded) != tc.want {
				t.Fatalf("JSON = %s, want %s", encoded, tc.want)
			}
		})
	}
}

func TestRecordJSONFieldOrderAndRoundTrip(t *testing.T) {
	input := Record{
		ArtifactID:   "artifact-1",
		RunID:        "run-1",
		NodeExecID:   "nodeexec-1",
		SessionID:    "session-1",
		ToolName:     "capture",
		Producer:     "tool:capture",
		Source:       "browser",
		Kind:         "screenshot",
		Role:         "evidence",
		Path:         ".agentx/capture.png",
		StorageRef:   "blob:1",
		URL:          "https://example.com",
		Digest:       "sha256:abc",
		MIMEType:     "image/png",
		Format:       "png",
		Bytes:        12,
		Summary:      "capture",
		Labels:       []string{"kind:screenshot"},
		MetadataJSON: `{"width":100}`,
		CreatedAt:    123,
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	want := `{"artifact_id":"artifact-1","run_id":"run-1","node_exec_id":"nodeexec-1","session_id":"session-1","tool_name":"capture","producer":"tool:capture","source":"browser","kind":"screenshot","role":"evidence","path":".agentx/capture.png","storage_ref":"blob:1","url":"https://example.com","digest":"sha256:abc","mime_type":"image/png","format":"png","bytes":12,"summary":"capture","labels":["kind:screenshot"],"metadata_json":"{\"width\":100}","created_at":123}`
	if string(encoded) != want {
		t.Fatalf("JSON = %s, want %s", encoded, want)
	}
	var decoded Record
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if !reflect.DeepEqual(decoded, input) {
		t.Fatalf("round trip = %#v, want %#v", decoded, input)
	}
}
