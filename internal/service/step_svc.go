package service

import (
	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/parse"
)

// ImportSteps 解析步骤文本并写入草稿（冻结草稿拒绝写入）。
func (s *Service) ImportSteps(draftID int64, text string) ([]*model.Step, error) {
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return nil, err
	}
	if frozen {
		return nil, model.ErrFrozenWrite
	}
	specs, err := parse.ParseSteps(text)
	if err != nil {
		return nil, err
	}
	if err := parse.ValidateLabelsUnique(specs); err != nil {
		return nil, err
	}
	var out []*model.Step
	for _, sp := range specs {
		st, err := s.store.CreateStep(&model.Step{
			DraftID:    draftID,
			Seq:        sp.Seq,
			Label:      sp.Label,
			Statement:  sp.Statement,
			Conclusion: sp.Conclusion,
			Status:     model.StepParsed,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	// 导入后草稿进入待复核
	_ = s.store.UpdateDraftStatus(draftID, model.DraftReviewing)
	return out, nil
}

// ListSteps 列出步骤。
func (s *Service) ListSteps(draftID int64) ([]*model.Step, error) {
	return s.store.ListSteps(draftID)
}

// ListStepPremises 查询某步骤的直接前提，并校验步骤属于草稿。
func (s *Service) ListStepPremises(draftID, stepID int64) ([]*model.PremiseEdge, error) {
	step, err := s.store.GetStep(stepID)
	if err != nil {
		return nil, err
	}
	if step.DraftID != draftID {
		return nil, model.ErrStepNotFound
	}
	return s.store.ListEdgesForStep(draftID, stepID)
}
