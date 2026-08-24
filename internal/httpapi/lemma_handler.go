package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleLemmas(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		ls, err := s.svc.ListLemmas(id)
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, ls)
	case http.MethodPost:
		var body struct {
			Name      string `json:"name"`
			Statement string `json:"statement"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		l, err := s.svc.CreateLemma(id, body.Name, body.Statement)
		if err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, l)
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (s *Server) handleReplaceLemma(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, 400, fmt.Errorf("bad path"))
		return
	}
	lid, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	var body struct {
		NewDraftID   int64  `json:"new_draft_id"`
		NewName      string `json:"new_name"`
		NewStatement string `json:"new_statement"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, err)
		return
	}
	l, err := s.svc.ReplaceLemma(lid, body.NewDraftID, body.NewName, body.NewStatement)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	writeJSON(w, 201, l)
}
