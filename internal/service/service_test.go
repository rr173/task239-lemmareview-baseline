package service

import (
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/store"
)

func newTestService(t *testing.T) (*Service, func()) {
	t.Helper()
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return New(st), func() { _ = st.Close() }
}

func TestAnalyzeCoversLinearChain(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	draft, err := svc.CreateDraft("linear", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, `1 A => A
2 B => B
3 C => C`)
	if err != nil {
		t.Fatal(err)
	}
	// 步骤2 依赖步骤1，步骤3 依赖步骤2
	if err := svc.AddPremise(draft.ID, steps[0].ID, "step", steps[1].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, steps[1].ID, "step", steps[2].ID, true); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CoveredSteps) != 3 {
		t.Fatalf("expected 3 covered, got %d (missing=%v cyclic=%v)", len(res.CoveredSteps), res.MissingSteps, res.CyclicSteps)
	}
	if len(res.Cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", res.Cycles)
	}
}

func TestAnalyzeDetectsCycle(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	draft, err := svc.CreateDraft("cycle", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, `1 X => X
2 Y => Y
3 Z => Z`)
	if err != nil {
		t.Fatal(err)
	}
	// 构造循环：1->2, 2->3, 3->1
	if err := svc.AddPremise(draft.ID, steps[0].ID, "step", steps[1].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, steps[1].ID, "step", steps[2].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, steps[2].ID, "step", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Cycles) == 0 {
		t.Fatal("expected at least one cycle detected")
	}
	if len(res.CyclicSteps) != 3 {
		t.Fatalf("expected 3 cyclic steps, got %v", res.CyclicSteps)
	}
}

func TestAddPremiseRejectsSelfReference(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, _ := svc.CreateDraft("self", "")
	steps, _ := svc.ImportSteps(draft.ID, `1 A => A`)
	if err := svc.AddPremise(draft.ID, steps[0].ID, "step", steps[0].ID, true); err == nil {
		t.Fatal("expected self-reference rejection")
	}
}

func TestFrozenDraftRejectsWrites(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, _ := svc.CreateDraft("fr", "")
	_, _ = svc.ImportSteps(draft.ID, `1 A => A`)
	if _, err := svc.FreezeVersion(draft.ID, "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ImportSteps(draft.ID, "2 B => B"); err == nil {
		t.Fatal("expected frozen draft write rejection")
	}
}
