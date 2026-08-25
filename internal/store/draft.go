package store

import (
	"database/sql"
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateDraft 新建草稿。
func (s *Store) CreateDraft(name, description string, status model.DraftStatus) (*model.Draft, error) {
	if name == "" {
		return nil, fmt.Errorf("draft name required")
	}
	t := nowStr()
	res, err := s.db.Exec(
		`INSERT INTO drafts(name, description, status, created_at, updated_at) VALUES(?,?,?,?,?)`,
		name, description, string(status), t, t)
	if err != nil {
		return nil, fmt.Errorf("insert draft: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetDraft(id)
}

// GetDraft 查询草稿。
func (s *Store) GetDraft(id int64) (*model.Draft, error) {
	row := s.db.QueryRow(`SELECT id,name,description,status,created_at,updated_at FROM drafts WHERE id=?`, id)
	d := &model.Draft{}
	var st, ca, ua string
	if err := row.Scan(&d.ID, &d.Name, &d.Description, &st, &ca, &ua); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrDraftNotFound
		}
		return nil, err
	}
	d.Status = model.DraftStatus(st)
	d.CreatedAt = parseTime(ca)
	d.UpdatedAt = parseTime(ua)
	return d, nil
}

// ListDrafts 列出全部草稿。
func (s *Store) ListDrafts() ([]*model.Draft, error) {
	rows, err := s.db.Query(`SELECT id,name,description,status,created_at,updated_at FROM drafts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Draft
	for rows.Next() {
		d := &model.Draft{}
		var st, ca, ua string
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &st, &ca, &ua); err != nil {
			return nil, err
		}
		d.Status = model.DraftStatus(st)
		d.CreatedAt = parseTime(ca)
		d.UpdatedAt = parseTime(ua)
		out = append(out, d)
	}
	return out, nil
}

// UpdateDraftStatus 更新草稿状态。
func (s *Store) UpdateDraftStatus(id int64, status model.DraftStatus) error {
	res, err := s.db.Exec(`UPDATE drafts SET status=?, updated_at=? WHERE id=?`, string(status), nowStr(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrDraftNotFound
	}
	return nil
}

// IsDraftFrozen 判断草稿是否已冻结（不可修改）。
func (s *Store) IsDraftFrozen(id int64) (bool, error) {
	d, err := s.GetDraft(id)
	if err != nil {
		return false, err
	}
	return d.Status == model.DraftFrozen, nil
}
