package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug02StepDetailStaysWithinDraft(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	one, err := svc.CreateDraft("one", "")
	if err != nil {
		t.Fatal(err)
	}
	two, err := svc.CreateDraft("two", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(one.ID, "1 private conclusion => secret")
	if err != nil {
		t.Fatal(err)
	}
	h := New(svc, ":0", "").Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/"+itoa(two.ID)+"/steps/"+itoa(steps[0].ID), bytes.NewReader(nil))
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("cross-draft step was exposed: status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func itoa(v int64) string {
	if v == 1 {
		return "1"
	}
	return "2"
}
