package service

import (
	"path/filepath"
	"sync"
	"testing"

	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug10ConcurrentFreezeHasSingleWinner(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	draft, err := svc.CreateDraft("concurrent freeze", "")
	if err != nil { t.Fatal(err) }
	if _, err := svc.ImportSteps(draft.ID, "1 immutable conclusion => result"); err != nil { t.Fatal(err) }
	const workers = 20
	start := make(chan struct{})
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.FreezeVersion(draft.ID, "shared candidate")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one concurrent freeze winner, got %d", successes)
	}
	versions, err := svc.ListVersions(draft.ID)
	if err != nil { t.Fatal(err) }
	if len(versions) != 1 {
		t.Fatalf("concurrent freeze created %d versions", len(versions))
	}
	got, err := st.GetDraft(draft.ID)
	if err != nil { t.Fatal(err) }
	if got.Status != model.DraftFrozen {
		t.Fatalf("draft did not remain frozen: %s", got.Status)
	}
}
