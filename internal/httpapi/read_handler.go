package httpapi

import (
	"fmt"
	"net/http"
)

// handleCoverage 提供只读覆盖查询，适合研究者反复刷新当前分析结果。
func (s *Server) handleCoverage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	res, err := s.svc.Analyze(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleGraph 返回草稿当前的步骤节点与依赖边，供前端或研究工具绘制证明图。
func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	id, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	steps, err := s.svc.ListSteps(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	edges, err := s.svc.Store().ListEdges(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"draft_id": id,
		"nodes":    steps,
		"edges":    edges,
	})
}

// handleStepPremises 返回单个步骤的直接前提，避免客户端重新过滤整张图。
func (s *Server) handleStepPremises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	draftID, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	stepID, err := parseLastID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	edges, err := s.svc.ListStepPremises(draftID, stepID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, edges)
}

// handleLemmaByID 返回指定草稿内的单条引理。
func (s *Server) handleLemmaByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}
	draftID, err := parseID(r.URL.Path, "/api/drafts/")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	lemmaID, err := parseLastID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	lemma, err := s.svc.GetLemma(draftID, lemmaID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, lemma)
}
