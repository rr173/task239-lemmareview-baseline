package cover

import (
	"testing"

	"task239-lemmareview/internal/model"
)

func TestAnalyzerPropagatesStepConclusions(t *testing.T) {
	steps := []*model.Step{{ID: 1}, {ID: 2}, {ID: 3}}
	edges := []*model.PremiseEdge{
		{FromKind: "step", FromID: 1, ToStepID: 2, Required: true},
		{FromKind: "step", FromID: 2, ToStepID: 3, Required: true},
	}
	res := New(steps, edges, nil).Analyze(nil)
	if len(res.CoveredSteps) != 3 || len(res.MissingSteps) != 0 {
		t.Fatalf("expected full propagation, got covered=%v missing=%v", res.CoveredSteps, res.MissingSteps)
	}
}

func TestStepPremiseLemmasIgnoresStepEdges(t *testing.T) {
	edges := []*model.PremiseEdge{
		{FromKind: "lemma", FromID: 9, ToStepID: 2},
		{FromKind: "step", FromID: 1, ToStepID: 2},
	}
	got := StepPremiseLemmas(edges)
	if len(got[2]) != 1 || got[2][0] != 9 {
		t.Fatalf("unexpected lemma premise map: %v", got)
	}
}

// TestAnalyzerOptionalPremiseDoesNotBlockCoverage 锁定：步骤仅有一个未满足的「可选」
// 引理前提时，该步骤仍判为 covered，而不落入 missing——可选前提不应阻断覆盖。
func TestAnalyzerOptionalPremiseDoesNotBlockCoverage(t *testing.T) {
	// 步骤 2 依赖：一个已满足的必修步骤前提（来自步骤1）+ 一个不可用的可选引理前提。
	// 该可选引理不在 availableLemmas 中，模拟「可选前提未满足」。
	steps := []*model.Step{{ID: 1}, {ID: 2}}
	edges := []*model.PremiseEdge{
		{FromKind: "step", FromID: 1, ToStepID: 2, Required: true},
		{FromKind: "lemma", FromID: 90, ToStepID: 2, Required: false},
	}
	res := New(steps, edges, nil).Analyze(nil)
	if len(res.CoveredSteps) != 2 || len(res.MissingSteps) != 0 {
		t.Fatalf("optional premise must not block coverage: covered=%v missing=%v", res.CoveredSteps, res.MissingSteps)
	}
}

// TestAnalyzerRequiredPremiseStillBlocksCoverage 对照组：同样的不可用引理前提若为必修，
// 则必须阻断该步骤覆盖并落入 missing，确保修复未误伤必修语义。
func TestAnalyzerRequiredPremiseStillBlocksCoverage(t *testing.T) {
	steps := []*model.Step{{ID: 1}, {ID: 2}}
	edges := []*model.PremiseEdge{
		{FromKind: "step", FromID: 1, ToStepID: 2, Required: true},
		{FromKind: "lemma", FromID: 90, ToStepID: 2, Required: true},
	}
	res := New(steps, edges, nil).Analyze(nil)
	if len(res.MissingSteps) != 1 || res.MissingSteps[0] != 2 {
		t.Fatalf("required unavailable premise must block coverage: missing=%v covered=%v", res.MissingSteps, res.CoveredSteps)
	}
}
