package service

import (
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateLemma 新建引理。
func (s *Service) CreateLemma(draftID int64, name, statement string) (*model.Lemma, error) {
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return nil, err
	}
	if frozen {
		return nil, model.ErrFrozenWrite
	}
	if name == "" {
		return nil, fmt.Errorf("lemma name required")
	}
	return s.store.CreateLemma(&model.Lemma{
		DraftID:   draftID,
		Name:      name,
		Statement: statement,
		Status:    model.LemmaCandidate,
	})
}

// ListLemmas 列出引理。
func (s *Service) ListLemmas(draftID int64) ([]*model.Lemma, error) {
	return s.store.ListLemmas(draftID)
}

// GetLemma 查询草稿中的单条引理。
func (s *Service) GetLemma(draftID, lemmaID int64) (*model.Lemma, error) {
	lemma, err := s.store.GetLemma(lemmaID)
	if err != nil {
		return nil, err
	}
	if lemma.DraftID != draftID {
		return nil, model.ErrLemmaNotFound
	}
	return lemma, nil
}

// SetLemmaStatus 更新引理可用性，仅允许草稿内的候选引理被标记为可用或失效。
func (s *Service) SetLemmaStatus(draftID, lemmaID int64, status model.LemmaStatus) error {
	lemma, err := s.GetLemma(draftID, lemmaID)
	if err != nil {
		return err
	}
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return err
	}
	if frozen {
		return model.ErrFrozenWrite
	}
	if lemma.Status != model.LemmaCandidate || (status != model.LemmaAvailable && status != model.LemmaInvalid) {
		return model.ErrInvalidStatus
	}
	return s.store.UpdateLemmaStatus(lemmaID, status)
}

// ReplaceLemma 替换引理：标记旧引理为 replaced，新引理为 available。
func (s *Service) ReplaceLemma(oldID, newDraftID int64, newName, newStatement string) (*model.Lemma, error) {
	old, err := s.store.GetLemma(oldID)
	if err != nil {
		return nil, err
	}
	frozen, err := s.store.IsDraftFrozen(old.DraftID)
	if err != nil {
		return nil, err
	}
	if frozen {
		return nil, model.ErrFrozenWrite
	}
	if newDraftID != old.DraftID || newName == "" {
		return nil, model.ErrInvalidStatus
	}
	return s.store.ReplaceLemma(oldID, newName, newStatement)
}

// AddPremise 新增前提依赖：校验非自指、已知引理、顺序合法。
func (s *Service) AddPremise(draftID, fromKindID int64, fromKind string, toStepID int64, required bool) error {
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return err
	}
	if frozen {
		return model.ErrFrozenWrite
	}
	return s.store.CreateEdge(&model.PremiseEdge{
		DraftID:  draftID,
		FromKind: fromKind,
		FromID:   fromKindID,
		ToStepID: toStepID,
		Required: required,
	})
}
