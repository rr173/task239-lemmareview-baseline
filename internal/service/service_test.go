package service

import (
	"path/filepath"
	"sync"
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

// TestConcurrentFreezeHasSingleWinner 验证并发冻结只有一个赢家：
// 20 个 goroutine 同时冻结同一草稿，最终恰好一个成功版本，其余返回 ErrFrozenWrite，
// 草稿保持 frozen，且版本表只有一行。
func TestConcurrentFreezeHasSingleWinner(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	draft, err := svc.CreateDraft("concurrent", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ImportSteps(draft.ID, "1 premise => conclusion"); err != nil {
		t.Fatal(err)
	}

	const n = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	frozenErrs := 0
	otherErrs := []string{}
	winVersionIDs := map[int64]bool{}
	wg.Add(n)
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start // 同时起跑，最大化竞态窗口
			v, err := svc.FreezeVersion(draft.ID, "v1")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				wins++
				winVersionIDs[v.ID] = true
				return
			}
			if err == model.ErrFrozenWrite {
				frozenErrs++
				return
			}
			otherErrs = append(otherErrs, err.Error())
		}()
	}
	close(start)
	wg.Wait()

	if wins != 1 {
		t.Fatalf("expected exactly 1 winner, got %d (frozenErrs=%d otherErrs=%v)", wins, frozenErrs, otherErrs)
	}
	if frozenErrs != n-1 {
		t.Fatalf("expected %d ErrFrozenWrite losers, got %d (otherErrs=%v)", n-1, frozenErrs, otherErrs)
	}
	if len(otherErrs) != 0 {
		t.Fatalf("unexpected errors: %v", otherErrs)
	}

	// 草稿最终必须保持冻结。
	d, err := svc.GetDraft(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != model.DraftFrozen {
		t.Fatalf("draft not frozen after concurrent freeze: %s", d.Status)
	}

	// 版本表只有一行成功版本。
	vs, err := svc.ListVersions(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 {
		t.Fatalf("expected exactly 1 persisted version, got %d", len(vs))
	}
	if !winVersionIDs[vs[0].ID] {
		t.Fatalf("persisted version id=%d was not reported by winner", vs[0].ID)
	}

	// 冻结后再次冻结应被拒绝（幂等保护）。
	if _, err := svc.FreezeVersion(draft.ID, "v2"); err != model.ErrFrozenWrite {
		t.Fatalf("expected ErrFrozenWrite on re-freeze, got %v", err)
	}
}
