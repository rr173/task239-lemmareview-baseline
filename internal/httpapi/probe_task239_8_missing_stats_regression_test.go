package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug08StatsRejectsMissingDraft(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	h := New(service.New(st), ":0", "").Handler()
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/api/drafts/999999/stats", nil))
	if resp.Code != http.StatusNotFound {
		t.Fatalf("missing draft stats returned success: status=%d body=%s", resp.Code, resp.Body.String())
	}
}
