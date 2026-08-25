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

func TestTask239Bug04PartialExemptionDoesNotCoverStep(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("partial exemption", "")
	if err != nil {
		t.Fatal(err)
	}
	steps, err := svc.ImportSteps(draft.ID, "1 needs two assumptions => result")
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.CreateLemma(draft.ID, "first", "first assumption")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateLemma(draft.ID, "second", "second assumption")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, first.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddPremise(draft.ID, second.ID, "lemma", steps[0].ID, true); err != nil {
		t.Fatal(err)
	}
	h := New(svc, ":0", "").Handler()
	body, _ := json.Marshal(map[string]any{"step_id": steps[0].ID, "lemma_id": second.ID, "reason": "second assumption accepted"})
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+strconv.FormatInt(draft.ID, 10)+"/exemptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to record exemption: status=%d body=%s", resp.Code, resp.Body.String())
	}
	analyzeReq := httptest.NewRequest(http.MethodPost, "/api/drafts/"+strconv.FormatInt(draft.ID, 10)+"/analyze", nil)
	analyzeResp := httptest.NewRecorder()
	h.ServeHTTP(analyzeResp, analyzeReq)
	if analyzeResp.Code != http.StatusOK {
		t.Fatalf("analysis failed: status=%d body=%s", analyzeResp.Code, analyzeResp.Body.String())
	}
	var result model.CoverageResult
	if err := json.Unmarshal(analyzeResp.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.ExemptedSteps) != 0 || len(result.MissingSteps) != 1 || result.MissingSteps[0] != steps[0].ID {
		t.Fatalf("partial exemption incorrectly covered step: %+v", result)
	}
	draftResp := httptest.NewRecorder()
	h.ServeHTTP(draftResp, httptest.NewRequest(http.MethodGet, "/api/drafts/"+strconv.FormatInt(draft.ID, 10), nil))
	var got model.Draft
	if err := json.Unmarshal(draftResp.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != model.DraftGap {
		t.Fatalf("draft status lost the remaining gap: %s", got.Status)
	}
}
