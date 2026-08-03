package brief

import (
	"encoding/json"
	"testing"

	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type testValidator struct{}

func (testValidator) ValidateSpec(agentxworkflow.Spec) error { return nil }

type testLowerer struct{}

func (testLowerer) LowerToolArguments(node agentxworkflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func testCoordinator(t *testing.T) *agentxpack.Coordinator {
	t.Helper()
	coordinator, err := agentxpack.NewCoordinator(testValidator{}, testLowerer{})
	if err != nil {
		t.Fatalf("new pack coordinator: %v", err)
	}
	return coordinator
}

func TestDefinitionValidatesAndMaterializes(t *testing.T) {
	def := Definition()
	if err := testCoordinator(t).ValidateDefinition(def); err != nil {
		t.Fatalf("validate definition: %v", err)
	}
	coordinator := testCoordinator(t)
	workflow, err := MaterializedDefaultWorkflow(coordinator)
	if err != nil {
		t.Fatalf("materialize workflow: %v", err)
	}
	if workflow.ID != DefaultWorkflow || workflow.Pack != PackID {
		t.Fatalf("unexpected workflow: %#v", workflow)
	}
	if suite, ok := def.EvalSuiteForWorkflow(DefaultWorkflow); !ok || suite.PassPath != "brief.passed" {
		t.Fatalf("expected default eval suite to gate on brief.passed, got ok=%v suite=%#v", ok, suite)
	}
}

func TestRegisterInto(t *testing.T) {
	reg, err := agentxpack.NewMemoryRegistry(testCoordinator(t))
	if err != nil {
		t.Fatalf("new pack registry: %v", err)
	}
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register pack: %v", err)
	}
	if _, ok := reg.Get(PackID); !ok {
		t.Fatalf("expected registered pack")
	}
}
