package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug07FrozenDraftRejectsLemmaReplacement(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("frozen lemma", "")
	if err != nil { t.Fatal(err) }
	lemma, err := svc.CreateLemma(draft.ID, "old", "old statement")
	if err != nil { t.Fatal(err) }
	if _, err := svc.FreezeVersion(draft.ID, "v1"); err != nil { t.Fatal(err) }
	h := New(svc, ":0", "").Handler()
	body, _ := json.Marshal(map[string]any{"new_draft_id": draft.ID, "new_name": "new", "new_statement": "new statement"})
	req := httptest.NewRequest(http.MethodPost, "/api/lemmas/"+strconv.FormatInt(lemma.ID, 10)+"/replace", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest { t.Fatalf("replacement modified frozen draft: status=%d body=%s", resp.Code, resp.Body.String()) }
	getResp := httptest.NewRecorder()
	h.ServeHTTP(getResp, httptest.NewRequest(http.MethodGet, "/api/drafts/"+strconv.FormatInt(draft.ID, 10)+"/lemmas/"+strconv.FormatInt(lemma.ID, 10), nil))
	var got model.Lemma
	if err := json.Unmarshal(getResp.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if got.Status != model.LemmaCandidate { t.Fatalf("frozen lemma status changed: %s", got.Status) }
}
