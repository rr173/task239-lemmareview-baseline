package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"task239-lemmareview/internal/model"
)

func (s *Server) handlePremises(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		es, err := s.svc.Store().ListEdges(id)
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, es)
	case http.MethodPost:
		var body struct {
			FromKind string `json:"from_kind"`
			FromID   int64  `json:"from_id"`
			ToStepID int64  `json:"to_step_id"`
			Required bool   `json:"required"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		if err := s.svc.Store().CreateEdge(&model.PremiseEdge{
			DraftID:  id,
			FromKind: body.FromKind,
			FromID:   body.FromID,
			ToStepID: body.ToStepID,
			Required: body.Required,
		}); err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, map[string]string{"status": "created"})
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	res, err := s.svc.Analyze(id)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	writeJSON(w, 200, res)
}

func (s *Server) handleExemptions(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		exs, err := s.svc.ListExemptions(id)
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, exs)
	case http.MethodPost:
		var body struct {
			StepID  int64  `json:"step_id"`
			LemmaID int64  `json:"lemma_id"`
			Reason  string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		if err := s.svc.AddExemption(id, body.StepID, body.LemmaID, body.Reason); err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, map[string]string{"status": "created"})
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}
