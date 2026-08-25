package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug06FrozenVersionRetainsExemptionEvidence(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("frozen evidence", "")
	if err != nil { t.Fatal(err) }
	steps, err := svc.ImportSteps(draft.ID, "1 relies on an accepted assumption => result")
	if err != nil { t.Fatal(err) }
	lemma, err := svc.CreateLemma(draft.ID, "accepted", "accepted assumption")
	if err != nil { t.Fatal(err) }
	if err := svc.AddPremise(draft.ID, lemma.ID, "lemma", steps[0].ID, true); err != nil { t.Fatal(err) }
	if err := svc.AddExemption(draft.ID, steps[0].ID, lemma.ID, "reviewer accepted this assumption"); err != nil { t.Fatal(err) }
	version, err := svc.FreezeVersion(draft.ID, "v1")
	if err != nil { t.Fatal(err) }
	h := New(svc, ":0", "").Handler()
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/api/versions/"+strconv.FormatInt(version.ID, 10), nil))
	if resp.Code != http.StatusOK { t.Fatalf("version lookup failed: status=%d body=%s", resp.Code, resp.Body.String()) }
	var got map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	want := "X[" + strconv.FormatInt(steps[0].ID, 10) + "|" + strconv.FormatInt(lemma.ID, 10) + "|reviewer accepted this assumption]"
	exemptions, _ := got["exemptions"].(string)
	snapshot, _ := got["snapshot"].(string)
	if !strings.Contains(exemptions, want) || !strings.Contains(snapshot, want) {
		t.Fatalf("frozen version lost exemption evidence: %+v", got)
	}
}
