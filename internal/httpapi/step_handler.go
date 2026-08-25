package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSteps(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		steps, err := s.svc.ListSteps(id)
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, steps)
	case http.MethodPost:
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		created, err := s.svc.ImportSteps(id, body.Text)
		if err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, created)
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (s *Server) handleStepByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, 400, fmt.Errorf("bad path"))
		return
	}
	sid, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		draftID, err := parseID(r.URL.Path, "/api/drafts/")
		if err != nil {
			writeError(w, 400, err)
			return
		}
		st, err := s.svc.GetStep(draftID, sid)
		if err != nil {
			writeError(w, 404, err)
			return
		}
		writeJSON(w, 200, st)
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}
