package service

import (
	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/version"
)

// FreezeVersion 冻结证明版本：绑定引理指纹与图快照，草稿置 frozen。
//
// 冻结通过 store.ClaimFreeze 在单个事务内完成条件状态翻转与版本插入，是并发冻结的
// 唯一同步点：多个并发调用中恰好一个赢家成功提交版本，其余返回 ErrFrozenWrite。
// 这避免了原先 IsDraftFrozen→CreateVersion→UpdateDraftStatus 之间的 check-then-act
// 竞态导致的多个成功版本。
func (s *Service) FreezeVersion(draftID int64, name string) (*model.ProofVersion, error) {
	lemmas, err := s.store.ListLemmas(draftID)
	if err != nil {
		return nil, err
	}
	steps, err := s.store.ListSteps(draftID)
	if err != nil {
		return nil, err
	}
	edges, err := s.store.ListEdges(draftID)
	if err != nil {
		return nil, err
	}
	v := version.Freeze(name, lemmas, steps, edges)
	v.DraftID = draftID
	created, err := s.store.ClaimFreeze(v)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// PublishShared 将已有冻结版本标记为共享（可复核分发）。
func (s *Service) PublishShared(versionID int64) error {
	v, err := s.store.GetVersion(versionID)
	if err != nil {
		return err
	}
	if v.Status != model.VersionFrozen {
		return model.ErrInvalidStatus
	}
	return s.store.UpdateVersionStatus(versionID, model.VersionShared)
}

// SupersedeVersion 将版本标记为被替代（新版本产生后旧版本失效）。
func (s *Service) SupersedeVersion(versionID int64) error {
	return s.store.UpdateVersionStatus(versionID, model.VersionSuperseded)
}

// ListVersions 列出版本。
func (s *Service) ListVersions(draftID int64) ([]*model.ProofVersion, error) {
	return s.store.ListVersions(draftID)
}
