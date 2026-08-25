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

// UpdateDraftStatus 按草稿状态机更新状态，冻结后不允许回退。
// 已冻结的草稿不得离开 frozen 状态（仅允许保持 frozen 的无操作写入），
// 否则调用方可经状态更新接口重新打开冻结草稿、绕过各写入用例的冻结保护。
func (s *Service) UpdateDraftStatus(id int64, next model.DraftStatus) error {
	d, err := s.store.GetDraft(id)
	if err != nil {
		return err
	}
	if d.Status == model.DraftFrozen && next != model.DraftFrozen {
		return model.ErrFrozenWrite
	}
	return s.store.UpdateDraftStatus(id, next)
}
