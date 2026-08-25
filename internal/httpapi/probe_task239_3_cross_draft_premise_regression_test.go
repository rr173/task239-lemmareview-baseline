package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug03CrossDraftPremiseIsRejected(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	sourceDraft, err := svc.CreateDraft("source", "")
	if err != nil {
		t.Fatal(err)
	}
	targetDraft, err := svc.CreateDraft("target", "")
	if err != nil {
		t.Fatal(err)
	}
	sourceSteps, err := svc.ImportSteps(sourceDraft.ID, "1 source conclusion => source")
	if err != nil {
		t.Fatal(err)
	}
	targetSteps, err := svc.ImportSteps(targetDraft.ID, "1 target conclusion => target")
	if err != nil {
		t.Fatal(err)
	}
	h := New(svc, ":0", "").Handler()
	body, _ := json.Marshal(map[string]any{
		"from_kind": "step",
		"from_id":   sourceSteps[0].ID,
		"to_step_id": targetSteps[0].ID,
		"required":  true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+strconv.FormatInt(targetDraft.ID, 10)+"/premises", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("cross-draft premise was accepted: status=%d body=%s", resp.Code, resp.Body.String())
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/drafts/"+strconv.FormatInt(targetDraft.ID, 10)+"/premises", nil)
	listResp := httptest.NewRecorder()
	h.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("premise listing failed: status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var edges []any
	if err := json.Unmarshal(listResp.Body.Bytes(), &edges); err != nil {
		t.Fatal(err)
	}
	if len(edges) != 0 {
		t.Fatalf("cross-draft premise remained visible: body=%s", listResp.Body.String())
	}
}
