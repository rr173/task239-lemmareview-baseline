package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"task239-lemmareview/internal/model"
)

func (s *Server) handleDrafts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ds, err := s.svc.ListDrafts()
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, ds)
	case http.MethodPost:
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		d, err := s.svc.CreateDraft(body.Name, body.Description)
		if err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, d)
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (s *Server) handleDraftByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		d, err := s.svc.GetDraft(id)
		if err != nil {
			writeError(w, 404, err)
			return
		}
		writeJSON(w, 200, d)
	case http.MethodPut:
		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		if err := s.svc.Store().UpdateDraftStatus(id, model.DraftStatus(body.Status)); err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 200, map[string]string{"status": "updated"})
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}
