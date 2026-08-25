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
