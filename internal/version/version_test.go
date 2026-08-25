package version

import (
	"strings"
	"testing"

	"task239-lemmareview/internal/model"
)

func TestGraphSnapshotIsIndependentOfInputOrder(t *testing.T) {
	a := []*model.Step{{ID: 2, Seq: 2, Label: "2"}, {ID: 1, Seq: 1, Label: "1"}}
	b := []*model.Step{{ID: 1, Seq: 1, Label: "1"}, {ID: 2, Seq: 2, Label: "2"}}
	edges := []*model.PremiseEdge{{FromKind: "step", FromID: 1, ToStepID: 2}}
	if GraphSnapshot(a, edges) != GraphSnapshot(b, edges) {
		t.Fatal("snapshot should be deterministic")
	}
}

func TestFreezeContainsLemmaFingerprintAndGraph(t *testing.T) {
	v := Freeze("v1", []*model.Lemma{{ID: 3, Status: model.LemmaAvailable}}, []*model.Step{{ID: 1, Label: "1"}}, nil)
	if v.Status != model.VersionFrozen || len(v.LemmaSet) != 64 || !strings.Contains(v.Snapshot, "S[1|") {
		t.Fatalf("unexpected frozen version: %#v", v)
	}
}
