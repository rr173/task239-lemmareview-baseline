package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		vs, err := s.svc.ListVersions(id)
		if err != nil {
			writeError(w, 500, err)
			return
		}
		writeJSON(w, 200, vs)
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, err)
			return
		}
		v, err := s.svc.FreezeVersion(id, body.Name)
		if err != nil {
			writeError(w, 400, err)
			return
		}
		writeJSON(w, 201, v)
	default:
		writeError(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (s *Server) handleVersionByID(w http.ResponseWriter, r *http.Request) {
	vid, err := parseLastID(r.URL.Path)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	v, err := s.svc.Store().GetVersion(vid)
	if err != nil {
		writeError(w, 404, err)
		return
	}
	writeJSON(w, 200, v)
}

func (s *Server) handleVersionShare(w http.ResponseWriter, r *http.Request) {
	vid, err := parseLastID(r.URL.Path)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	if err := s.svc.PublishShared(vid); err != nil {
		writeError(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "shared"})
}

func (s *Server) handleVersionSupersede(w http.ResponseWriter, r *http.Request) {
	vid, err := parseLastID(r.URL.Path)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	if err := s.svc.SupersedeVersion(vid); err != nil {
		writeError(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "superseded"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, 400, err)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	steps, _ := s.svc.ListSteps(id)
	lemmas, _ := s.svc.ListLemmas(id)
	edges, _ := s.svc.Store().ListEdges(id)
	stat := map[string]int{
		"steps":  len(steps),
		"lemmas": len(lemmas),
		"edges":  len(edges),
	}
	writeJSON(w, 200, stat)
}
