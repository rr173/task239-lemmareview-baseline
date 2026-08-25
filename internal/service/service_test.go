package service

import (
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/model"
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

func TestAnalyzeAppliesCompleteLemmaExemption(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("exemption", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 needs an external assumption => result")
	if err != nil {
		t.Fatal(err)
	}
	lemma, err := svc.CreateLemma(draft.ID, "external", "assumption")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, lemma.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddExemption(draft.ID, steps[0].ID, lemma.ID, "accepted by the author"); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.ExemptedSteps) != 1 || len(res.MissingSteps) != 0 {
		t.Fatalf("expected one exempted step and no missing steps, got exempted=%v missing=%v", res.ExemptedSteps, res.MissingSteps)
	}
}

func TestImportStepsRollsBackDuplicateBatch(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("atomic", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ImportSteps(draft.ID, "1 original => ok"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ImportSteps(draft.ID, "2 inserted first => bad\n1 duplicate => should rollback"); err == nil {
		t.Fatal("expected duplicate label failure")
	}
	steps, err := svc.ListSteps(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Label != "1" {
		t.Fatalf("failed batch left partial steps: %#v", steps)
	}
}

func TestDraftStatusCannotLeaveFrozenThroughService(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("frozen-status", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ImportSteps(draft.ID, "1 A => A"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.FreezeVersion(draft.ID, "v1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.UpdateDraftStatus(draft.ID, model.DraftEditing); err != model.ErrFrozenWrite {
		t.Fatalf("expected frozen write rejection, got %v", err)
	}
}

func TestReplaceLemmaMigratesPremiseEdges(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("replace", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 depends on lemma => result")
	if err != nil {
		t.Fatal(err)
	}
	oldLemma, err := svc.CreateLemma(draft.ID, "old", "old statement")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, oldLemma.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	newLemma, err := svc.ReplaceLemma(oldLemma.ID, draft.ID, "new", "new statement")
	if err != nil {
		t.Fatal(err)
	}
	edges, err := svc.Store().ListEdges(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 1 || edges[0].FromID != newLemma.ID {
		t.Fatalf("premise edge was not migrated: %#v", edges)
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil || len(res.CoveredSteps) != 1 {
		t.Fatalf("replacement did not preserve coverage: result=%#v err=%v", res, err)
	}
}

// TestAnalyzeOptionalLemmaDoesNotBlockPublishable 锁定端到端修复：步骤仅依赖一个未满足的
// 「可选」引理前提（引理保持 candidate、未置 available）时，分析应判该步骤为 covered、
// missing 为空，并把草稿置为 publishable——可选前提不应阻断覆盖或草稿发布。
// 该用例同时覆盖 Required 字段经 store 写读往返后仍为可选（验证 ListEdges 不再把可选当必修）。
func TestAnalyzeOptionalLemmaDoesNotBlockPublishable(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("optional-premise", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 base => A\n2 from A with optional lemma => B")
	if err != nil {
		t.Fatal(err)
	}
	// 步骤2 依赖步骤1（必修，已覆盖）+ 一个可选引理前提（引理不置 available，模拟未满足）
	if err := svc.AddPremise(draft.ID, steps[0].ID, "step", steps[1].ID, true); err != nil {
		t.Fatal(err)
	}
	optionalLemma, err := svc.CreateLemma(draft.ID, "optional", "an optional assumption")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, optionalLemma.ID, "lemma", steps[1].ID, false); err != nil {
		t.Fatal(err)
	}
	// 验证 store 读回的 Required 仍为可选（往返保真，不应被 ListEdges 改写成必修）
	edges, err := svc.Store().ListEdges(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range edges {
		if e.FromKind == "lemma" && e.FromID == optionalLemma.ID && e.Required {
			t.Fatalf("optional premise edge lost its optional flag after round-trip: %#v", e)
		}
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MissingSteps) != 0 {
		t.Fatalf("optional premise must not yield missing steps: %v", res.MissingSteps)
	}
	if len(res.CoveredSteps) != 2 {
		t.Fatalf("expected both steps covered, got covered=%v", res.CoveredSteps)
	}
	got, err := svc.GetDraft(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.DraftPublishable {
		t.Fatalf("expected draft publishable, got status=%s", got.Status)
	}
}

// TestAnalyzeRequiredLemmaGapKeepsGap 对照组：同样一个未满足引理前提若为必修，
// 草稿必须留在 gap 状态，确保修复未放宽必修前提的缺口判定。
func TestAnalyzeRequiredLemmaGapKeepsGap(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	draft, err := svc.CreateDraft("required-premise", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 depends on lemma => result")
	if err != nil {
		t.Fatal(err)
	}
	lemma, err := svc.CreateLemma(draft.ID, "missing", "an unavailable lemma")
	if err != nil {
		t.Fatal(err)
	}
	// 引理保持 candidate（未 available），必修前提未满足
	if err := svc.AddPremise(draft.ID, lemma.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MissingSteps) != 1 || res.MissingSteps[0] != steps[0].ID {
		t.Fatalf("required unavailable premise must yield a missing step: missing=%v", res.MissingSteps)
	}
	got, err := svc.GetDraft(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.DraftGap {
		t.Fatalf("expected draft gap, got status=%s", got.Status)
	}
}
