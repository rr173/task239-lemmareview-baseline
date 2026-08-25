package service

import (
	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/version"
)

// FreezeVersion 冻结证明版本：绑定引理指纹与图快照，草稿置 frozen。
func (s *Service) FreezeVersion(draftID int64, name string) (*model.ProofVersion, error) {
	frozen, err := s.store.IsDraftFrozen(draftID)
	if err != nil {
		return nil, err
	}
	if frozen {
		return nil, model.ErrFrozenWrite
	}
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
	created, err := s.store.CreateVersion(v)
	if err != nil {
		return nil, err
	}
	if err := s.store.UpdateDraftStatus(draftID, model.DraftFrozen); err != nil {
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
// 仅允许替代已共享（shared）的版本：未共享的冻结版本仍处于评审生命周期内，
// 不得跳过 shared 状态直接替代，否则会绕过复核分发环节。
func (s *Service) SupersedeVersion(versionID int64) error {
	v, err := s.store.GetVersion(versionID)
	if err != nil {
		return err
	}
	if v.Status != model.VersionShared {
		return model.ErrInvalidStatus
	}
	return s.store.UpdateVersionStatus(versionID, model.VersionSuperseded)
}

// ListVersions 列出版本。
func (s *Service) ListVersions(draftID int64) ([]*model.ProofVersion, error) {
	return s.store.ListVersions(draftID)
}
