package graph

import (
	"testing"

	"task239-lemmareview/internal/model"
)

func TestDetectCyclesReturnsClosedCycle(t *testing.T) {
	steps := []*model.Step{{ID: 1}, {ID: 2}, {ID: 3}}
	edges := []*model.PremiseEdge{
		{FromKind: "step", FromID: 1, ToStepID: 2},
		{FromKind: "step", FromID: 2, ToStepID: 3},
		{FromKind: "step", FromID: 3, ToStepID: 1},
	}
	cycles, err := Build(steps, edges).DetectCycles()
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) != 1 || cycles[0][0] != cycles[0][len(cycles[0])-1] {
		t.Fatalf("expected one closed cycle, got %v", cycles)
	}
}

func TestDirectCyclicStepsFindsSelfReference(t *testing.T) {
	g := Build([]*model.Step{{ID: 7}}, []*model.PremiseEdge{{FromKind: "step", FromID: 7, ToStepID: 7}})
	if !g.DirectCyclicSteps()[7] {
		t.Fatal("expected self-referencing step to be cyclic")
	}
}
