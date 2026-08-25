package store

import (
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateVersion 新增证明版本快照。
func (s *Store) CreateVersion(v *model.ProofVersion) (*model.ProofVersion, error) {
	t := nowStr()
	if v.Status == "" {
		v.Status = model.VersionDraft
	}
	res, err := s.db.Exec(
		`INSERT INTO proof_versions(draft_id,name,status,lemma_set,snapshot,created_at) VALUES(?,?,?,?,?,?)`,
		v.DraftID, v.Name, string(v.Status), v.LemmaSet, v.Snapshot, t)
	if err != nil {
		return nil, fmtErrInsertVersion(err)
	}
	id, _ := res.LastInsertId()
	return s.GetVersion(id)
}

// ClaimFreeze 原子地冻结草稿并写入证明版本快照，是并发冻结的唯一同步点。
//
// 在单个事务内执行条件状态翻转：仅当草稿当前未冻结时，UPDATE 才会命中一行并随之插入
// 版本记录。SQLite 单写连接保证该事务对其他事务是串行执行的，因此 20 个并发调用者中
// 恰好有一个 RowsAffected=1 成为赢家并提交版本；其余调用者命中 0 行，回滚后返回
// ErrFrozenWrite。这消除了原先 IsDraftFrozen→CreateVersion→UpdateDraftStatus 之间的
// check-then-act 竞态，保证并发冻结只有一个赢家且最终草稿保持冻结。
//
// 返回新创建的版本（赢家）或 model.ErrFrozenWrite（输家）。
func (s *Store) ClaimFreeze(v *model.ProofVersion) (*model.ProofVersion, error) {
	t := nowStr()
	if v.Status == "" {
		v.Status = model.VersionFrozen
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin freeze: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // 提交后回滚为空操作

	// CAS：仅当草稿尚未冻结时翻转状态。locked 标记本次调用是否抢到冻结权。
	res, err := tx.Exec(
		`UPDATE drafts SET status=?, updated_at=? WHERE id=? AND status<>?`,
		string(model.DraftFrozen), t, v.DraftID, string(model.DraftFrozen))
	if err != nil {
		return nil, fmt.Errorf("claim freeze: %w", err)
	}
	locked, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if locked == 0 {
		// 已被其他并发赢家冻结，本次调用为输家。
		return nil, model.ErrFrozenWrite
	}

	// 赢家在同一事务内写入版本快照，保证草稿冻结与版本落库原子一致。
	if v.Status == "" {
		v.Status = model.VersionFrozen
	}
	ins, err := tx.Exec(
		`INSERT INTO proof_versions(draft_id,name,status,lemma_set,snapshot,created_at) VALUES(?,?,?,?,?,?)`,
		v.DraftID, v.Name, string(v.Status), v.LemmaSet, v.Snapshot, t)
	if err != nil {
		return nil, fmtErrInsertVersion(err)
	}
	id, err := ins.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit freeze: %w", err)
	}
	return s.GetVersion(id)
}

// GetVersion 查询版本。
func (s *Store) GetVersion(id int64) (*model.ProofVersion, error) {
	row := s.db.QueryRow(`SELECT id,draft_id,name,status,lemma_set,snapshot,created_at FROM proof_versions WHERE id=?`, id)
	v := &model.ProofVersion{}
	var st, ca string
	if err := row.Scan(&v.ID, &v.DraftID, &v.Name, &st, &v.LemmaSet, &v.Snapshot, &ca); err != nil {
		return nil, err
	}
	v.Status = model.VersionStatus(st)
	v.CreatedAt = parseTime(ca)
	return v, nil
}

// ListVersions 列出草稿版本。
func (s *Store) ListVersions(draftID int64) ([]*model.ProofVersion, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,name,status,lemma_set,snapshot,created_at FROM proof_versions WHERE draft_id=? ORDER BY id`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ProofVersion
	for rows.Next() {
		v := &model.ProofVersion{}
		var st, ca string
		if err := rows.Scan(&v.ID, &v.DraftID, &v.Name, &st, &v.LemmaSet, &v.Snapshot, &ca); err != nil {
			return nil, err
		}
		v.Status = model.VersionStatus(st)
		v.CreatedAt = parseTime(ca)
		out = append(out, v)
	}
	return out, nil
}

// UpdateVersionStatus 更新版本状态（冻结/替代）。
func (s *Store) UpdateVersionStatus(id int64, status model.VersionStatus) error {
	res, err := s.db.Exec(`UPDATE proof_versions SET status=? WHERE id=?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errVersionNotFound
	}
	return nil
}

var errVersionNotFound = fmtErrVersion("version not found")

type versionErr string

func (e versionErr) Error() string { return string(e) }

func fmtErrInsertVersion(err error) error {
	return fmt.Errorf("insert version: %w", err)
}

func fmtErrVersion(msg string) error { return versionErr(msg) }
