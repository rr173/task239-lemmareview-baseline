package adjudicate

import (
	"testing"

	"task239-lemmareview/internal/model"
)

func TestApplyExemptionsRequiresEveryLemma(t *testing.T) {
	engine := New([]*model.Exemption{{StepID: 4, LemmaID: 10, Reason: "accepted"}})
	missing, exempted := engine.ApplyExemptions([]int64{4, 5}, map[int64][]int64{
		4: {10, 11},
		5: {12},
	})
	if len(exempted) != 0 || len(missing) != 2 {
		t.Fatalf("partial exemptions must remain missing: missing=%v exempted=%v", missing, exempted)
	}
	engine = New([]*model.Exemption{{StepID: 4, LemmaID: 10}, {StepID: 4, LemmaID: 11}})
	missing, exempted = engine.ApplyExemptions([]int64{4}, map[int64][]int64{4: {10, 11}})
	if len(missing) != 0 || len(exempted) != 1 || exempted[0] != 4 {
		t.Fatalf("complete exemption not applied: missing=%v exempted=%v", missing, exempted)
	}
}
