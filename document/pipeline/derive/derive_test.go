package derive_test

import (
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/derive"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"reflect"
	"testing"
)

// helper to build spec/result quickly
func makeSpecWithDerived(chapter string, fields map[string]string) *configs.DocSpec {
	ch := configs.ChapterSpec{Key: chapter}
	for k, f := range fields {
		ch.Fields = append(ch.Fields, configs.FieldSpec{Key: k, Type: "number", DerivedFormula: f})
	}
	return &configs.DocSpec{Chapters: []configs.ChapterSpec{ch}}
}

func TestDerived_MissingDependency(t *testing.T) {
	// A.X depends on B.Y (missing)
	spec := &configs.DocSpec{Chapters: []configs.ChapterSpec{
		{Key: "A", Fields: []configs.FieldSpec{{Key: "X", Type: "number", DerivedFormula: "B.Y + 1"}}},
	}}
	res := &types.DocumentResult{Chapters: map[string]*types.ChapterResult{"A": {Key: "A", Fields: map[string]types.FieldResult{}}}}

	derive.EvaluateDerived(spec, res)

	if len(res.DerivedDiagnostics) == 0 {
		t.Fatalf("expected diagnostics for missing deps")
	}
	d := res.DerivedDiagnostics[0]
	if d.ID != "A.X" || d.Cycle {
		t.Fatalf("unexpected diag: %+v", d)
	}
	// Missing should include B.Y
	if got := d.MissingDeps; !reflect.DeepEqual(got, []string{"B.Y"}) {
		t.Fatalf("missing deps mismatch: %v", got)
	}
}

func TestDerived_BlockedByPending(t *testing.T) {
	// A.X depends on B.Y; B.Y depends on C.W (missing)
	spec := &configs.DocSpec{Chapters: []configs.ChapterSpec{
		{Key: "A", Fields: []configs.FieldSpec{{Key: "X", Type: "number", DerivedFormula: "B.Y + 1"}}},
		{Key: "B", Fields: []configs.FieldSpec{{Key: "Y", Type: "number", DerivedFormula: "C.W + 1"}}},
	}}
	res := &types.DocumentResult{Chapters: map[string]*types.ChapterResult{"A": {Key: "A", Fields: map[string]types.FieldResult{}}, "B": {Key: "B", Fields: map[string]types.FieldResult{}}}}

	derive.EvaluateDerived(spec, res)

	if len(res.DerivedDiagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(res.DerivedDiagnostics))
	}
	// find A.X
	var dx, dy types.DerivedDiagnostic
	for _, d := range res.DerivedDiagnostics {
		if d.ID == "A.X" {
			dx = d
		}
		if d.ID == "B.Y" {
			dy = d
		}
	}
	if len(dx.BlockedBy) == 0 || dx.BlockedBy[0] != "B.Y" {
		t.Fatalf("A.X should be blocked by B.Y: %+v", dx)
	}
	if got := dy.MissingDeps; !reflect.DeepEqual(got, []string{"C.W"}) {
		t.Fatalf("B.Y missing deps mismatch: %v", got)
	}
}

func TestDerived_Cycle(t *testing.T) {
	// A.X <-> B.Y cycle
	spec := &configs.DocSpec{Chapters: []configs.ChapterSpec{
		{Key: "A", Fields: []configs.FieldSpec{{Key: "X", Type: "number", DerivedFormula: "B.Y + 1"}}},
		{Key: "B", Fields: []configs.FieldSpec{{Key: "Y", Type: "number", DerivedFormula: "A.X + 1"}}},
	}}
	res := &types.DocumentResult{Chapters: map[string]*types.ChapterResult{"A": {Key: "A", Fields: map[string]types.FieldResult{}}, "B": {Key: "B", Fields: map[string]types.FieldResult{}}}}

	derive.EvaluateDerived(spec, res)

	if len(res.DerivedDiagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics for cycle, got %d", len(res.DerivedDiagnostics))
	}
	for _, d := range res.DerivedDiagnostics {
		if !d.Cycle {
			t.Fatalf("expected cycle=true: %+v", d)
		}
	}
}

func TestDerivedDiagnosticsSortedByID(t *testing.T) {
	spec := &configs.DocSpec{Chapters: []configs.ChapterSpec{
		{Key: "B", Fields: []configs.FieldSpec{{Key: "Y", Type: "number", DerivedFormula: "Missing.Value + 1"}}},
		{Key: "A", Fields: []configs.FieldSpec{{Key: "X", Type: "number", DerivedFormula: "Missing.Value + 1"}}},
	}}
	res := &types.DocumentResult{Chapters: map[string]*types.ChapterResult{
		"A": {Key: "A", Fields: map[string]types.FieldResult{}},
		"B": {Key: "B", Fields: map[string]types.FieldResult{}},
	}}

	derive.EvaluateDerived(spec, res)

	if len(res.DerivedDiagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(res.DerivedDiagnostics))
	}
	if res.DerivedDiagnostics[0].ID != "A.X" || res.DerivedDiagnostics[1].ID != "B.Y" {
		t.Fatalf("diagnostics should be sorted by ID: %+v", res.DerivedDiagnostics)
	}
}
