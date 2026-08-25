package httpapi

import (
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

func TestTask239Bug09OptionalPremiseDoesNotBlockCoverage(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("optional premise", "")
	if err != nil { t.Fatal(err) }
	steps, err := svc.ImportSteps(draft.ID, "1 has an optional assumption => result")
	if err != nil { t.Fatal(err) }
	lemma, err := svc.CreateLemma(draft.ID, "optional", "not required")
	if err != nil { t.Fatal(err) }
	if err := svc.AddPremise(draft.ID, lemma.ID, "lemma", steps[0].ID, false); err != nil { t.Fatal(err) }
	h := New(svc, ":0", "").Handler()
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/api/drafts/"+strconv.FormatInt(draft.ID, 10)+"/analyze", nil))
	if resp.Code != http.StatusOK { t.Fatalf("analysis failed: status=%d body=%s", resp.Code, resp.Body.String()) }
	var result model.CoverageResult
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil { t.Fatal(err) }
	if len(result.MissingSteps) != 0 || len(result.CoveredSteps) != 1 || result.CoveredSteps[0] != steps[0].ID { t.Fatalf("optional premise blocked coverage: %+v", result) }
	getResp := httptest.NewRecorder()
	h.ServeHTTP(getResp, httptest.NewRequest(http.MethodGet, "/api/drafts/"+strconv.FormatInt(draft.ID, 10), nil))
	var got model.Draft
	if err := json.Unmarshal(getResp.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if got.Status != model.DraftPublishable { t.Fatalf("optional premise left draft unpublished: %s", got.Status) }
}
