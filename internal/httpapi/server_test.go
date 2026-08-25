package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestReadHandlersExposeGraphAndLemma(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("api", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 premise => conclusion")
	if err != nil {
		t.Fatal(err)
	}
	lemma, err := svc.CreateLemma(draft.ID, "L", "lemma statement")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, lemma.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	h := New(svc, ":0", "").Handler()

	graphReq := httptest.NewRequest(http.MethodGet, "/api/drafts/1/graph", nil)
	graphResp := httptest.NewRecorder()
	h.ServeHTTP(graphResp, graphReq)
	if graphResp.Code != http.StatusOK {
		t.Fatalf("graph status=%d body=%s", graphResp.Code, graphResp.Body.String())
	}
	var graphBody struct {
		Nodes []any `json:"nodes"`
		Edges []any `json:"edges"`
	}
	if err := json.Unmarshal(graphResp.Body.Bytes(), &graphBody); err != nil || len(graphBody.Nodes) != 1 || len(graphBody.Edges) != 1 {
		t.Fatalf("unexpected graph response: %s", graphResp.Body.String())
	}

	lemmaReq := httptest.NewRequest(http.MethodGet, "/api/drafts/1/lemmas/1", nil)
	lemmaResp := httptest.NewRecorder()
	h.ServeHTTP(lemmaResp, lemmaReq)
	if lemmaResp.Code != http.StatusOK || !json.Valid(lemmaResp.Body.Bytes()) {
		t.Fatalf("unexpected lemma response: status=%d body=%s", lemmaResp.Code, lemmaResp.Body.String())
	}
}
