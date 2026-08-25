package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func TestTask239Bug01FrozenDraftCannotBeReopened(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := New(service.New(st), ":0", "").Handler()
	draftResp := serveJSON(h, http.MethodPost, "/api/drafts", `{"name":"frozen","description":""}`)
	if draftResp.Code != http.StatusCreated {
		t.Fatalf("create draft: %d %s", draftResp.Code, draftResp.Body.String())
	}
	if got := serveJSON(h, http.MethodPost, "/api/drafts/1/steps", `{"text":"1 initial => result"}`).Code; got != http.StatusCreated {
		t.Fatalf("import initial step: %d", got)
	}
	if got := serveJSON(h, http.MethodPost, "/api/drafts/1/versions", `{"name":"v1"}`).Code; got != http.StatusCreated {
		t.Fatalf("freeze draft: %d", got)
	}
	statusResp := serveJSON(h, http.MethodPut, "/api/drafts/1", `{"status":"editing"}`)
	if statusResp.Code == http.StatusOK {
		t.Fatalf("frozen draft was reopened: %s", statusResp.Body.String())
	}
	writeResp := serveJSON(h, http.MethodPost, "/api/drafts/1/steps", `{"text":"2 forbidden => write"}`)
	if writeResp.Code == http.StatusCreated {
		t.Fatalf("frozen draft accepted a write")
	}
	var draft map[string]any
	if err := json.NewDecoder(serveJSON(h, http.MethodGet, "/api/drafts/1", "").Body).Decode(&draft); err != nil {
		t.Fatal(err)
	}
	if draft["status"] != "frozen" {
		t.Fatalf("draft left frozen state: %#v", draft)
	}
}

func serveJSON(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	return resp
}
