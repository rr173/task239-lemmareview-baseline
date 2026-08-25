package service

import (
	"task239-lemmareview/internal/adjudicate"
	"task239-lemmareview/internal/cover"
	"task239-lemmareview/internal/graph"
	"task239-lemmareview/internal/model"
)

// Analyze 执行完整覆盖分析：构建图、检测循环、传播前提。
func (s *Service) Analyze(draftID int64) (*model.CoverageResult, error) {
	steps, err := s.store.ListSteps(draftID)
	if err != nil {
		return nil, err
	}
	edges, err := s.store.ListEdges(draftID)
	if err != nil {
		return nil, err
	}
	lemmas, err := s.store.ListLemmas(draftID)
	if err != nil {
		return nil, err
	}

	g := graph.Build(steps, edges)
	cycles, err := g.DetectCycles()
	if err != nil {
		return nil, err
	}
	cyclicSet := make(map[int64]bool)
	for _, c := range cycles {
		for _, id := range c {
			cyclicSet[id] = true
		}
	}

	analyzer := cover.New(steps, edges, lemmas)
	res := analyzer.Analyze(cyclicSet)
	res.Cycles = cycles
	exms, err := s.store.ListExemptions(draftID)
	if err != nil {
		return nil, err
	}
	engine := adjudicate.New(exms)
	stillMissing, exempted := engine.ApplyExemptions(res.MissingSteps, cover.StepPremiseLemmas(edges))
	res.MissingSteps = stillMissing
	res.ExemptedSteps = exempted
	res.CoveredSteps = append(res.CoveredSteps, exempted...)

	// 同步步骤状态
	for _, sid := range res.CoveredSteps {
		if err := s.store.SetStepStatus(sid, model.StepCovered); err != nil {
			return nil, err
		}
	}
	for _, sid := range res.MissingSteps {
		if err := s.store.SetStepStatus(sid, model.StepMissing); err != nil {
			return nil, err
		}
	}
	for _, sid := range res.CyclicSteps {
		if err := s.store.SetStepStatus(sid, model.StepCyclic); err != nil {
			return nil, err
		}
	}

	// 草稿缺口/可发布状态：仅当残留未豁免的缺失步骤或存在循环时才判为缺口；
	// 存在前提边本身不应阻断一个已完全覆盖的草稿进入可发布状态。
	if len(res.MissingSteps) > 0 || len(res.CyclicSteps) > 0 {
		if err := s.store.UpdateDraftStatus(draftID, model.DraftGap); err != nil {
			return nil, err
		}
	} else {
		if err := s.store.UpdateDraftStatus(draftID, model.DraftPublishable); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// AddExemption 新增假设豁免。
func (s *Service) AddExemption(draftID, stepID, lemmaID int64, reason string) error {
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return err
	}
	if frozen {
		return model.ErrFrozenWrite
	}
	step, err := s.store.GetStep(stepID)
	if err != nil {
		return err
	}
	if step.DraftID != draftID {
		return model.ErrStepNotFound
	}
	if lemmaID != 0 {
		lemma, err := s.store.GetLemma(lemmaID)
		if err != nil || lemma.DraftID != draftID {
			return model.ErrUnknownLemma
		}
	}
	return s.store.CreateExemption(&model.Exemption{
		DraftID: draftID,
		StepID:  stepID,
		LemmaID: lemmaID,
		Reason:  reason,
	})
}

// ListExemptions 列出豁免。
func (s *Service) ListExemptions(draftID int64) ([]*model.Exemption, error) {
	return s.store.ListExemptions(draftID)
}

// GetExemption 查询单条假设豁免，并确保它属于指定草稿。
func (s *Service) GetExemption(draftID, exemptionID int64) (*model.Exemption, error) {
	ex, err := s.store.GetExemption(exemptionID)
	if err != nil {
		return nil, err
	}
	if ex.DraftID != draftID {
		return nil, model.ErrStepNotFound
	}
	return ex, nil
}
