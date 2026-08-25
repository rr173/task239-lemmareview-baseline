// Package httpapi 暴露 HTTP 接口：路由前缀 /api，覆盖草稿、步骤、引理、
// 前提边、覆盖分析、豁免、版本与自检等能力（≥20 个端点）。
// 各资源 handler 拆到 draft_handler.go / step_handler.go / lemma_handler.go /
// analyze_handler.go / version_handler.go；本文件含服务结构、路由注册与通用辅助。
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"task239-lemmareview/internal/service"
)

// Server HTTP 服务。
type Server struct {
	svc  *service.Service
	addr string
	db   string
}

// New 构造服务。
func New(svc *service.Service, addr, dbPath string) *Server {
	return &Server{svc: svc, addr: addr, db: dbPath}
}

// Handler 返回注册好路由的 http.Handler。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	// 草稿
	mux.HandleFunc("/api/drafts", s.handleDrafts)     // GET 列表 / POST 新建
	mux.HandleFunc("/api/drafts/", s.handleDraftByID) // GET / PUT 状态
	// 步骤
	mux.HandleFunc("/api/drafts/{id}/steps", s.handleSteps)          // GET 列表 / POST 文本导入
	mux.HandleFunc("/api/drafts/{id}/steps/{sid}", s.handleStepByID) // GET
	// 引理
	mux.HandleFunc("/api/drafts/{id}/lemmas", s.handleLemmas)         // GET / POST
	mux.HandleFunc("/api/lemmas/{lid}/status", s.handleLemmaStatus)   // PUT 可用性
	mux.HandleFunc("/api/lemmas/{lid}/replace", s.handleReplaceLemma) // POST
	// 前提边
	mux.HandleFunc("/api/drafts/{id}/premises", s.handlePremises) // GET / POST
	// 覆盖分析
	mux.HandleFunc("/api/drafts/{id}/analyze", s.handleAnalyze)                   // POST
	mux.HandleFunc("/api/drafts/{id}/coverage", s.handleCoverage)                 // GET 当前覆盖结果
	mux.HandleFunc("/api/drafts/{id}/graph", s.handleGraph)                       // GET 步骤与依赖图
	mux.HandleFunc("/api/drafts/{id}/steps/{sid}/premises", s.handleStepPremises) // GET 步骤前提
	mux.HandleFunc("/api/drafts/{id}/lemmas/{lid}", s.handleLemmaByID)            // GET 单条引理
	// 豁免
	mux.HandleFunc("/api/drafts/{id}/exemptions", s.handleExemptions) // GET / POST
	// 版本
	mux.HandleFunc("/api/drafts/{id}/versions", s.handleVersions)             // GET / POST 冻结
	mux.HandleFunc("/api/versions/{vid}", s.handleVersionByID)                // GET
	mux.HandleFunc("/api/versions/{vid}/share", s.handleVersionShare)         // POST 共享
	mux.HandleFunc("/api/versions/{vid}/supersede", s.handleVersionSupersede) // POST 替代
	// 统计与自检
	mux.HandleFunc("/api/drafts/{id}/stats", s.handleStats) // GET
	mux.HandleFunc("/api/selfcheck", s.handleSelfCheck)     // GET
	return mux
}

// Run 启动 HTTP 服务（长驻）。
func (s *Server) Run() error {
	fmt.Printf("lemmareview listening on %s (db=%s)\n", s.addr, s.db)
	return http.ListenAndServe(s.addr, s.Handler())
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

// parseID 从路径中解析首个路径段 ID。
func parseID(path string, prefix string) (int64, error) {
	rest := strings.TrimPrefix(path, prefix)
	rest = strings.Trim(rest, "/")
	idx := strings.Index(rest, "/")
	if idx >= 0 {
		rest = rest[:idx]
	}
	return strconv.ParseInt(rest, 10, 64)
}

// parseLastID 解析路径末尾的 ID（如 /api/versions/{vid}）。
func parseLastID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	last := parts[len(parts)-1]
	return strconv.ParseInt(last, 10, 64)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) handleSelfCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "module": "task239-lemmareview"})
}
