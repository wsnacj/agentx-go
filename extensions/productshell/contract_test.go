package productshell

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	agentxcases "github.com/wsnacj/agentx-go/runtime/cases"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"

	"github.com/wsnacj/agentx-go/extensions/pack"
)

func TestShellBindingCodecDeepClone(t *testing.T) {
	nested := map[string]any{"items": []any{map[string]any{"name": "before"}}}
	payload := ShellBindingOptionPayload{PackID: " pack ", CaseInput: nested, Persist: true}
	binding, hasValues, persist, err := DecodeShellBindingOption(payload)
	if err != nil || !hasValues || !persist {
		t.Fatalf("DecodeShellBindingOption() = (%#v, %v, %v, %v)", binding, hasValues, persist, err)
	}
	nested["items"].([]any)[0].(map[string]any)["name"] = "after"
	got := binding.CaseInput["items"].([]any)[0].(map[string]any)["name"]
	if got != "before" {
		t.Fatalf("decoded binding aliases caller memory: got %v", got)
	}
	if binding.PackID != "pack" {
		t.Fatalf("PackID = %q, want pack", binding.PackID)
	}
}

func TestProjectInputOptionsPreservesPriorityAndSemanticMaps(t *testing.T) {
	raw := map[string]any{
		"auto_case_binding": "false",
		"requested_skill_semantics": []map[string]any{{
			"name": " Review ", "execution_context": "fork", "allowed_tools": []string{" Read ", "read"},
		}},
		"session_input": map[string]any{"source": "typed"},
		"sessionInput":  map[string]any{"source": "fallback"},
		"workflow_id":   "workflow-a",
	}
	options, _, sessionInput, _, residual := ProjectInputOptions(raw)
	if options.AutoCaseBinding == nil || *options.AutoCaseBinding {
		t.Fatalf("AutoCaseBinding = %#v, want false", options.AutoCaseBinding)
	}
	if got := options.RequestedSkillSemantics; len(got) != 1 || got[0].Name != "review" || got[0].ExecutionContext != "fork" || !reflect.DeepEqual(got[0].AllowedTools, []string{"read"}) {
		t.Fatalf("RequestedSkillSemantics = %#v", got)
	}
	if sessionInput["source"] != "typed" {
		t.Fatalf("session input priority changed: %#v", sessionInput)
	}
	if residual["workflow_id"] != "workflow-a" {
		t.Fatalf("residual = %#v", residual)
	}
}

func TestApplyCommandDispatchDoesNotAliasRequestedSkills(t *testing.T) {
	backing := make([]string, 1, 4)
	backing[0] = "existing"
	input := Input{ShellOptions: InputShellOptions{RequestedSkills: backing}}
	out := ApplyCommandDispatch(input, &PreparedCommandDispatch{Matched: true, Skill: "new"})
	if !reflect.DeepEqual(out.ShellOptions.RequestedSkills, []string{"existing", "new"}) {
		t.Fatalf("RequestedSkills = %#v", out.ShellOptions.RequestedSkills)
	}
	if got := backing[:2][1]; got != "" {
		t.Fatalf("caller backing array mutated: %q", got)
	}
}

func TestParseRequestedSkillsDoesNotInferProductPolicy(t *testing.T) {
	requested, message := ParseRequestedSkills(Input{ProductShell: "coding", UserMessage: "implement this"})
	if len(requested) != 0 || message != "implement this" {
		t.Fatalf("ParseRequestedSkills() = (%#v, %q), want explicit-only result", requested, message)
	}
}

func TestPreparationPipelineOrderAndContext(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "identity")
	var calls []string
	record := func(name string) { calls = append(calls, name) }
	checkContext := func(callCtx context.Context) {
		if callCtx != ctx || callCtx.Value(contextKey{}) != "identity" {
			t.Fatalf("pipeline did not forward context identity")
		}
	}
	spec := agentxworkflow.Spec{ID: "workflow-a"}
	binding := pack.Binding{PackID: "pack-a", CaseType: "case-a", WorkflowID: "workflow-a"}
	effectiveCase := agentxcases.Case{ID: "case-1"}
	rt := PreparationRuntimeFuncs{
		ApplyInputCaseFn: func(input Input) Input { record("apply-input"); return input },
		ResolveShellBindingFn: func(callCtx context.Context, _ string, _ Input) (*PreparedShellBinding, error) {
			checkContext(callCtx)
			record("resolve-shell")
			return &PreparedShellBinding{Matched: true}, nil
		},
		ApplyShellBindingFn: func(input Input, _ *PreparedShellBinding) (Input, error) { record("apply-shell"); return input, nil },
		ResolveCommandDispatchFn: func(callCtx context.Context, _ Input) (*PreparedCommandDispatch, error) {
			checkContext(callCtx)
			record("resolve-command")
			return &PreparedCommandDispatch{Matched: true}, nil
		},
		ApplyCommandDispatchFn:       func(input Input, _ *PreparedCommandDispatch) Input { record("apply-command"); return input },
		ParseRequestedSkillsFn:       func(Input) ([]string, string) { record("parse-skills"); return []string{"one"}, "message" },
		ShouldAttemptPackSelectionFn: func(Input) bool { record("should-pack"); return true },
		ResolvePackSelectionFn: func(callCtx context.Context, _ Input, _ string) (*PreparedPackSelection, error) {
			checkContext(callCtx)
			record("resolve-pack")
			return &PreparedPackSelection{Matched: true, Applied: true, Binding: WorkflowBinding{PackID: "pack-a", CaseType: "case-a"}}, nil
		},
		ApplyPackSelectionFn:       func(input Input, _ *PreparedPackSelection) Input { record("apply-pack"); return input },
		ShouldAttemptCaseBindingFn: func(Input) bool { record("should-case"); return true },
		ResolveCandidateCaseBindingFn: func(Input, *PreparedPackSelection) (pack.Binding, bool, error) {
			record("resolve-candidate")
			return binding, true, nil
		},
		ResolveCaseBindingDraftFn: func(callCtx context.Context, _ Input, _ string, _ pack.Binding, _ int) (*PreparedCaseBinding, error) {
			checkContext(callCtx)
			record("resolve-case")
			return &PreparedCaseBinding{Applied: true}, nil
		},
		MergeCaseBindingMetricsFn: func(metrics CaseBindingMetrics, _ *PreparedCaseBinding) CaseBindingMetrics {
			record("merge-case")
			return metrics
		},
		ApplyCaseBindingFn: func(input Input, _ *PreparedCaseBinding) Input { record("apply-case"); return input },
		ResolveWorkflowFn: func(Input) (ResolvedWorkflow, error) {
			record("resolve-workflow")
			return ResolvedWorkflow{Spec: &spec, PackBinding: &binding}, nil
		},
		ResolveEffectiveCaseFn: func(_ string, _ Input, _ string, _ *agentxworkflow.Spec, _ *pack.Binding) (*agentxcases.Case, error) {
			record("resolve-effective")
			return &effectiveCase, nil
		},
		ValidateEffectiveCaseFn: func(*pack.Binding, *agentxcases.Case) error { record("validate-effective"); return nil },
		ApplyEffectiveCaseFn:    func(input Input, _ agentxcases.Case) Input { record("apply-effective"); return input },
		FinalizeShellBindingMetricsFn: func(metrics ShellBindingMetrics, _ Input, _ *pack.Binding, _ *PreparedShellBinding) ShellBindingMetrics {
			record("finalize-shell")
			return metrics
		},
		PackSelectionMetricsFromFn: func(*PreparedPackSelection) PackSelectionMetrics {
			record("pack-metrics")
			return PackSelectionMetrics{}
		},
	}
	_, err := NewPreparationPipeline(rt).PrepareWithInput(ctx, PrepareInput{SessionID: "session", LLMTaskTimeoutMs: 10})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apply-input", "resolve-shell", "apply-shell", "resolve-command", "apply-command", "parse-skills", "should-pack", "resolve-pack", "apply-pack", "should-case", "resolve-candidate", "resolve-case", "merge-case", "apply-case", "resolve-workflow", "resolve-effective", "validate-effective", "apply-effective", "finalize-shell", "pack-metrics"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("stage order:\n got %#v\nwant %#v", calls, want)
	}
}

func TestPreparationPipelineStopsAtFirstError(t *testing.T) {
	sentinel := errors.New("command failed")
	var calls []string
	rt := PreparationRuntimeFuncs{
		ApplyInputCaseFn: func(input Input) Input { calls = append(calls, "apply-input"); return input },
		ResolveShellBindingFn: func(context.Context, string, Input) (*PreparedShellBinding, error) {
			calls = append(calls, "resolve-shell")
			return nil, nil
		},
		ResolveCommandDispatchFn: func(context.Context, Input) (*PreparedCommandDispatch, error) {
			calls = append(calls, "resolve-command")
			return nil, sentinel
		},
		ParseRequestedSkillsFn: func(Input) ([]string, string) { calls = append(calls, "unexpected"); return nil, "" },
	}
	_, err := NewPreparationPipeline(rt).Prepare(context.Background(), "session", Input{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel identity", err)
	}
	if want := []string{"apply-input", "resolve-shell", "resolve-command"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestPreparationPipelineForwardsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rt := PreparationRuntimeFuncs{ResolveShellBindingFn: func(callCtx context.Context, _ string, _ Input) (*PreparedShellBinding, error) {
		return nil, callCtx.Err()
	}}
	_, err := NewPreparationPipeline(rt).Prepare(ctx, "session", Input{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestPreparationPipelineStopsAfterValidationError(t *testing.T) {
	sentinel := errors.New("validation failed")
	var calls []string
	spec := agentxworkflow.Spec{ID: "workflow"}
	effective := agentxcases.Case{ID: "case"}
	rt := PreparationRuntimeFuncs{
		ResolveWorkflowFn: func(Input) (ResolvedWorkflow, error) {
			calls = append(calls, "resolve-workflow")
			return ResolvedWorkflow{Spec: &spec}, nil
		},
		ResolveEffectiveCaseFn: func(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error) {
			calls = append(calls, "resolve-effective")
			return &effective, nil
		},
		ValidateEffectiveCaseFn: func(*pack.Binding, *agentxcases.Case) error {
			calls = append(calls, "validate-effective")
			return sentinel
		},
		ApplyEffectiveCaseFn: func(input Input, _ agentxcases.Case) Input {
			calls = append(calls, "unexpected-apply")
			return input
		},
		PackSelectionMetricsFromFn: func(*PreparedPackSelection) PackSelectionMetrics {
			calls = append(calls, "unexpected-metrics")
			return PackSelectionMetrics{}
		},
	}
	_, err := NewPreparationPipeline(rt).Prepare(context.Background(), "session", Input{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel identity", err)
	}
	want := []string{"resolve-workflow", "resolve-effective", "validate-effective"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestPreparationPipelinePreservesLaterStageFirstError(t *testing.T) {
	sentinel := errors.New("stage failed")
	stages := []string{
		"apply-shell",
		"resolve-pack",
		"resolve-candidate",
		"resolve-case",
		"resolve-workflow",
		"resolve-effective",
		"validate-effective",
	}
	for _, stop := range stages {
		t.Run(stop, func(t *testing.T) {
			var calls []string
			record := func(stage string) { calls = append(calls, stage) }
			fail := func(stage string) error {
				if stage == stop {
					return sentinel
				}
				return nil
			}
			spec := agentxworkflow.Spec{ID: "workflow"}
			binding := pack.Binding{PackID: "pack", CaseType: "case", WorkflowID: "workflow"}
			effective := agentxcases.Case{ID: "case-id"}
			rt := PreparationRuntimeFuncs{
				ResolveShellBindingFn: func(context.Context, string, Input) (*PreparedShellBinding, error) {
					record("resolve-shell")
					return &PreparedShellBinding{Matched: true}, nil
				},
				ApplyShellBindingFn: func(input Input, _ *PreparedShellBinding) (Input, error) {
					record("apply-shell")
					return input, fail("apply-shell")
				},
				ShouldAttemptPackSelectionFn: func(Input) bool { return true },
				ResolvePackSelectionFn: func(context.Context, Input, string) (*PreparedPackSelection, error) {
					record("resolve-pack")
					return &PreparedPackSelection{}, fail("resolve-pack")
				},
				ShouldAttemptCaseBindingFn: func(Input) bool { return true },
				ResolveCandidateCaseBindingFn: func(Input, *PreparedPackSelection) (pack.Binding, bool, error) {
					record("resolve-candidate")
					return binding, true, fail("resolve-candidate")
				},
				ResolveCaseBindingDraftFn: func(context.Context, Input, string, pack.Binding, int) (*PreparedCaseBinding, error) {
					record("resolve-case")
					return &PreparedCaseBinding{}, fail("resolve-case")
				},
				ResolveWorkflowFn: func(Input) (ResolvedWorkflow, error) {
					record("resolve-workflow")
					return ResolvedWorkflow{Spec: &spec, PackBinding: &binding}, fail("resolve-workflow")
				},
				ResolveEffectiveCaseFn: func(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error) {
					record("resolve-effective")
					return &effective, fail("resolve-effective")
				},
				ValidateEffectiveCaseFn: func(*pack.Binding, *agentxcases.Case) error {
					record("validate-effective")
					return fail("validate-effective")
				},
				PackSelectionMetricsFromFn: func(*PreparedPackSelection) PackSelectionMetrics {
					record("unexpected-after-error")
					return PackSelectionMetrics{}
				},
			}
			_, err := NewPreparationPipeline(rt).Prepare(context.Background(), "session", Input{})
			if !errors.Is(err, sentinel) {
				t.Fatalf("error = %v, want sentinel identity", err)
			}
			if got := calls[len(calls)-1]; got != stop {
				t.Fatalf("last call = %q, want %q; calls=%#v", got, stop, calls)
			}
			for _, call := range calls {
				if call == "unexpected-after-error" {
					t.Fatalf("pipeline continued after %s: %#v", stop, calls)
				}
			}
		})
	}
}

func TestPreparationPipelineAdapterFailuresStopImmediately(t *testing.T) {
	sentinel := errors.New("adapter failed")
	tests := []struct {
		name      string
		configure func(*PreparationRuntimeFuncs, *[]string)
		want      []string
	}{
		{
			name: "pack selection",
			configure: func(rt *PreparationRuntimeFuncs, trace *[]string) {
				rt.ShouldAttemptPackSelectionFn = func(Input) bool { return true }
				rt.ResolvePackSelectionFn = func(context.Context, Input, string) (*PreparedPackSelection, error) {
					*trace = append(*trace, "resolve-pack")
					return nil, sentinel
				}
				rt.ShouldAttemptCaseBindingFn = func(Input) bool { *trace = append(*trace, "unexpected"); return false }
			},
			want: []string{"resolve-pack"},
		},
		{
			name: "candidate binding",
			configure: func(rt *PreparationRuntimeFuncs, trace *[]string) {
				rt.ShouldAttemptCaseBindingFn = func(Input) bool { return true }
				rt.ResolveCandidateCaseBindingFn = func(Input, *PreparedPackSelection) (pack.Binding, bool, error) {
					*trace = append(*trace, "resolve-candidate")
					return pack.Binding{}, false, sentinel
				}
				rt.ResolveWorkflowFn = func(Input) (ResolvedWorkflow, error) {
					*trace = append(*trace, "unexpected")
					return ResolvedWorkflow{}, nil
				}
			},
			want: []string{"resolve-candidate"},
		},
		{
			name: "case draft",
			configure: func(rt *PreparationRuntimeFuncs, trace *[]string) {
				rt.ShouldAttemptCaseBindingFn = func(Input) bool { return true }
				rt.ResolveCandidateCaseBindingFn = func(Input, *PreparedPackSelection) (pack.Binding, bool, error) { return pack.Binding{}, true, nil }
				rt.ResolveCaseBindingDraftFn = func(context.Context, Input, string, pack.Binding, int) (*PreparedCaseBinding, error) {
					*trace = append(*trace, "resolve-draft")
					return nil, sentinel
				}
				rt.ResolveWorkflowFn = func(Input) (ResolvedWorkflow, error) {
					*trace = append(*trace, "unexpected")
					return ResolvedWorkflow{}, nil
				}
			},
			want: []string{"resolve-draft"},
		},
		{
			name: "workflow",
			configure: func(rt *PreparationRuntimeFuncs, trace *[]string) {
				rt.ResolveWorkflowFn = func(Input) (ResolvedWorkflow, error) {
					*trace = append(*trace, "resolve-workflow")
					return ResolvedWorkflow{}, sentinel
				}
				rt.ResolveEffectiveCaseFn = func(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error) {
					*trace = append(*trace, "unexpected")
					return nil, nil
				}
			},
			want: []string{"resolve-workflow"},
		},
		{
			name: "effective case",
			configure: func(rt *PreparationRuntimeFuncs, trace *[]string) {
				rt.ResolveEffectiveCaseFn = func(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error) {
					*trace = append(*trace, "resolve-effective")
					return nil, sentinel
				}
				rt.ValidateEffectiveCaseFn = func(*pack.Binding, *agentxcases.Case) error { *trace = append(*trace, "unexpected"); return nil }
			},
			want: []string{"resolve-effective"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var trace []string
			rt := PreparationRuntimeFuncs{}
			test.configure(&rt, &trace)
			result, err := NewPreparationPipeline(rt).Prepare(context.Background(), "session", Input{})
			if !errors.Is(err, sentinel) {
				t.Fatalf("error = %v, want sentinel identity", err)
			}
			if !reflect.DeepEqual(result, PrepareResult{}) {
				t.Fatalf("result = %#v, want zero value", result)
			}
			if !reflect.DeepEqual(trace, test.want) {
				t.Fatalf("trace = %#v, want %#v", trace, test.want)
			}
		})
	}
}

func TestPreparationPipelineNilPassThrough(t *testing.T) {
	input := Input{UserMessage: "unchanged", Options: map[string]any{"key": "value"}}
	var nilPipeline *PreparationPipeline
	for name, pipeline := range map[string]*PreparationPipeline{
		"nil pipeline": nilPipeline,
		"nil runtime":  NewPreparationPipeline(nil),
	} {
		t.Run(name, func(t *testing.T) {
			result, err := pipeline.Prepare(context.Background(), "session", input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Input, input) || result.Workflow == nil || !reflect.DeepEqual(*result.Workflow, ResolvedWorkflow{}) {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestPreparationRuntimeFuncsNilMergePreservesMetrics(t *testing.T) {
	want := CaseBindingMetrics{Attempted: true, Source: "existing"}
	got := (PreparationRuntimeFuncs{}).MergeCaseBindingMetrics(want, &PreparedCaseBinding{Metrics: CaseBindingMetrics{Source: "replacement"}})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeCaseBindingMetrics() = %#v, want %#v", got, want)
	}
}

func TestLoadSessionShellBindingMetaJSONPreservesZeroValue(t *testing.T) {
	binding, ok, err := LoadSessionShellBindingMetaJSON(`{"agentx_shell_binding":{}}`)
	if err != nil || ok {
		t.Fatalf("LoadSessionShellBindingMetaJSON() = (%#v, %v, %v)", binding, ok, err)
	}
	if !reflect.DeepEqual(binding, ShellBinding{}) {
		t.Fatalf("empty binding = %#v, want exact zero value", binding)
	}
	if keys := sortedShellBindingMapKeys(nil); keys != nil {
		t.Fatalf("sortedShellBindingMapKeys(nil) = %#v, want nil", keys)
	}
}

func TestMergeSessionShellBindingMetaJSONPreservesOtherMetadata(t *testing.T) {
	encoded, err := MergeSessionShellBindingMetaJSON(`{"other":{"value":1}}`, ShellBinding{PackID: "pack-a"})
	if err != nil {
		t.Fatal(err)
	}
	binding, ok, err := LoadSessionShellBindingMetaJSON(encoded)
	if err != nil || !ok || binding.PackID != "pack-a" {
		t.Fatalf("round trip = (%#v, %v, %v)", binding, ok, err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(encoded), &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["other"]; !ok {
		t.Fatalf("unrelated metadata lost: %s", encoded)
	}
	removed, err := MergeSessionShellBindingMetaJSON(encoded, ShellBinding{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := LoadSessionShellBindingMetaJSON(removed); err != nil || ok {
		t.Fatalf("removed binding still present: ok=%v err=%v payload=%s", ok, err, removed)
	}
}

func TestWorkflowBindingPrefersCaseThenTypedThenOptions(t *testing.T) {
	input := Input{
		Case:   &agentxcases.Case{PackID: "case-pack", Type: "case-type", WorkflowID: "case-workflow"},
		PackID: "typed-pack", CaseType: "typed-type", PackWorkflow: "typed-workflow",
		Options: map[string]any{"pack_id": "option-pack", "case_type": "option-type", "workflow_id": "option-workflow"},
	}
	binding, ok, err := ResolveWorkflowBinding(input)
	if err != nil || !ok {
		t.Fatalf("ResolveWorkflowBinding() = (%#v, %v, %v)", binding, ok, err)
	}
	want := WorkflowBinding{PackID: "case-pack", CaseType: "case-type", WorkflowID: "case-workflow"}
	if !reflect.DeepEqual(binding, want) {
		t.Fatalf("binding = %#v, want %#v", binding, want)
	}
}

func TestWorkflowResolutionRawOptIn(t *testing.T) {
	spec := agentxworkflow.Spec{ID: "raw"}
	rt := WorkflowResolutionRuntime{}
	if _, err := rt.ResolveExplicitWorkflow(Input{}, spec); err == nil || err.Error() != "agentx: explicit raw workflow requires raw-workflow opt-in" {
		t.Fatalf("error = %v", err)
	}
	resolved, err := rt.ResolveExplicitWorkflow(Input{RawWorkflowOptIn: true}, spec)
	if err != nil || resolved.Spec == nil || resolved.Spec.ID != "raw" || resolved.PackBinding != nil {
		t.Fatalf("ResolveExplicitWorkflow() = (%#v, %v)", resolved, err)
	}
}

func TestWorkflowResolutionSemanticEquivalence(t *testing.T) {
	registered := agentxworkflow.Spec{
		ID: "workflow-a", Pack: "pack-a", CaseTypes: []string{"case-a"}, EntryNode: "node-a",
		Nodes: []agentxworkflow.NodeSpec{{ID: "node-a", Kind: agentxworkflow.NodeTool, Title: "registered title", Description: "registered description", Config: map[string]any{"tool": "lookup"}}},
	}
	binding := pack.Binding{PackID: "pack-a", CaseType: "case-a", WorkflowID: "workflow-a", Workflow: registered}
	rt := WorkflowResolutionRuntime{
		HasRegisteredPackFn:             func(string) bool { return true },
		ResolveExplicitPackBindingFn:    func(string, string, string) (pack.Binding, bool, error) { return binding, true, nil },
		MaterializeRegisteredWorkflowFn: func(pack.Binding) (agentxworkflow.Spec, error) { return registered, nil },
	}
	explicit := registered
	explicit.Nodes = append([]agentxworkflow.NodeSpec(nil), registered.Nodes...)
	explicit.Title = "caller title"
	explicit.Description = "caller description"
	explicit.RouteHints = []string{"documentation-only-difference"}
	explicit.Nodes[0].Title = "caller node title"
	resolved, err := rt.ResolveExplicitWorkflow(Input{CaseType: "case-a"}, explicit)
	if err != nil || resolved.PackBinding == nil || resolved.Spec == nil {
		t.Fatalf("equivalent workflow rejected: (%#v, %v)", resolved, err)
	}
	drift := explicit
	drift.Nodes = append([]agentxworkflow.NodeSpec(nil), explicit.Nodes...)
	drift.Nodes[0].Config = map[string]any{"tool": "different"}
	if _, err := rt.ResolveExplicitWorkflow(Input{CaseType: "case-a"}, drift); err == nil {
		t.Fatal("execution semantic drift was accepted")
	}
}
