// Package service 编排层：串联解析、依赖图、覆盖分析、豁免裁决与版本冻结，
// 对 store 做事务性封装，并向 HTTP 层暴露业务用例。
// 各业务用例按资源拆到 draft_svc.go / step_svc.go / lemma_svc.go /
// analyze_svc.go / version_svc.go；本文件仅含服务结构与构造。
package service

import (
	"task239-lemmareview/internal/store"
)

// Service 业务编排服务。
type Service struct {
	store *store.Store
}

// New 构造服务。
func New(st *store.Store) *Service {
	return &Service{store: st}
}

// Store 暴露底层存储（供特殊用例与 HTTP 层直接读写）。
func (s *Service) Store() *store.Store { return s.store }
