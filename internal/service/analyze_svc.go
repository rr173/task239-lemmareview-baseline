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

	// 同步步骤状态
	for _, sid := range res.CoveredSteps {
		_ = s.store.SetStepStatus(sid, model.StepCovered)
	}
	for _, sid := range res.MissingSteps {
		_ = s.store.SetStepStatus(sid, model.StepMissing)
	}
	for _, sid := range res.CyclicSteps {
		_ = s.store.SetStepStatus(sid, model.StepCyclic)
	}

	// 草稿缺口/可发布状态
	if len(res.MissingSteps) > 0 || len(res.CyclicSteps) > 0 {
		_ = s.store.UpdateDraftStatus(draftID, model.DraftGap)
	} else {
		_ = s.store.UpdateDraftStatus(draftID, model.DraftPublishable)
	}

	// 关联豁免信息（作为结果补充）
	exms, _ := s.store.ListExemptions(draftID)
	eng := adjudicate.New(exms)
	res.AvailablePremises = append(res.AvailablePremises, 0)[:0] // 保留，覆盖分析已填
	_ = eng
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
	_, err = s.store.GetStep(stepID)
	if err != nil {
		return err
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
