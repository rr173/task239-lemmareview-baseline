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

func TestTask239Bug05UnsharedVersionCannotBeSuperseded(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	draft, err := svc.CreateDraft("version lifecycle", "")
	if err != nil { t.Fatal(err) }
	if _, err := svc.ImportSteps(draft.ID, "1 stable conclusion => result"); err != nil { t.Fatal(err) }
	version, err := svc.FreezeVersion(draft.ID, "v1")
	if err != nil { t.Fatal(err) }
	h := New(svc, ":0", "").Handler()
	path := "/api/versions/" + strconv.FormatInt(version.ID, 10) + "/supersede"
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, path, nil))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unshared version was superseded: status=%d body=%s", resp.Code, resp.Body.String())
	}
	getResp := httptest.NewRecorder()
	h.ServeHTTP(getResp, httptest.NewRequest(http.MethodGet, "/api/versions/"+strconv.FormatInt(version.ID, 10), nil))
	var got model.ProofVersion
	if err := json.Unmarshal(getResp.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if got.Status != model.VersionFrozen {
		t.Fatalf("version status changed after rejected transition: %s", got.Status)
	}
}
