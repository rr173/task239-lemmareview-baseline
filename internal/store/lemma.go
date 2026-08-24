package store

import (
	"database/sql"
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateLemma 新建引理。
func (s *Store) CreateLemma(lemma *model.Lemma) (*model.Lemma, error) {
	if lemma.DraftID == 0 {
		return nil, errDraftRequired
	}
	t := nowStr()
	if lemma.Status == "" {
		lemma.Status = model.LemmaCandidate
	}
	res, err := s.db.Exec(
		`INSERT INTO lemmas(draft_id,name,statement,status,created_at) VALUES(?,?,?,?,?)`,
		lemma.DraftID, lemma.Name, lemma.Statement, string(lemma.Status), t)
	if err != nil {
		return nil, fmtErrInsertLemma(err)
	}
	id, _ := res.LastInsertId()
	return s.GetLemma(id)
}

// GetLemma 查询引理。
func (s *Store) GetLemma(id int64) (*model.Lemma, error) {
	row := s.db.QueryRow(`SELECT id,draft_id,name,statement,status,created_at FROM lemmas WHERE id=?`, id)
	l := &model.Lemma{}
	var st, ca string
	if err := row.Scan(&l.ID, &l.DraftID, &l.Name, &l.Statement, &st, &ca); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrLemmaNotFound
		}
		return nil, err
	}
	l.Status = model.LemmaStatus(st)
	l.CreatedAt = parseTime(ca)
	return l, nil
}

// ListLemmas 列出草稿内引理。
func (s *Store) ListLemmas(draftID int64) ([]*model.Lemma, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,name,statement,status,created_at FROM lemmas WHERE draft_id=? ORDER BY id`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Lemma
	for rows.Next() {
		l := &model.Lemma{}
		var st, ca string
		if err := rows.Scan(&l.ID, &l.DraftID, &l.Name, &l.Statement, &st, &ca); err != nil {
			return nil, err
		}
		l.Status = model.LemmaStatus(st)
		l.CreatedAt = parseTime(ca)
		out = append(out, l)
	}
	return out, nil
}

// UpdateLemmaStatus 更新引理状态（替换/失效）。
func (s *Store) UpdateLemmaStatus(id int64, status model.LemmaStatus) error {
	res, err := s.db.Exec(`UPDATE lemmas SET status=? WHERE id=?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrLemmaNotFound
	}
	return nil
}

func fmtErrInsertLemma(err error) error {
	return fmt.Errorf("insert lemma: %w", err)
}
