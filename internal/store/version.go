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
