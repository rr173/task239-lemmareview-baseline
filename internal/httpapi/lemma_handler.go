package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"task239-lemmareview/internal/model"
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
	// 经由 service 层而非直接调用 store：ReplaceLemma 在 store 写入前会校验
	// 草稿是否冻结，冻结时拒绝替换并保持旧引理状态不变。
	l, err := s.svc.ReplaceLemma(lid, body.NewDraftID, body.NewName, body.NewStatement)
	if err != nil {
		writeError(w, 400, err)
		return
	}
	writeJSON(w, 201, l)
}

func (s *Server) handleLemmaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("bad path"))
		return
	}
	lid, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		DraftID int64             `json:"draft_id"`
		Status  model.LemmaStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.SetLemmaStatus(body.DraftID, lid, body.Status); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(body.Status)})
}
