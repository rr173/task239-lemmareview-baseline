package service

import (
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateDraft 新建草稿（仅编辑中状态可创建）。
func (s *Service) CreateDraft(name, description string) (*model.Draft, error) {
	if name == "" {
		return nil, fmt.Errorf("draft name required")
	}
	return s.store.CreateDraft(name, description, model.DraftEditing)
}

// GetDraft 查询草稿。
func (s *Service) GetDraft(id int64) (*model.Draft, error) {
	return s.store.GetDraft(id)
}

// ListDrafts 列出草稿。
func (s *Service) ListDrafts() ([]*model.Draft, error) {
	return s.store.ListDrafts()
}
