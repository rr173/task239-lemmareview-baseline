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

// ReplaceLemma 原子替换引理，并把证明图中的前提边迁移到新引理。
func (s *Store) ReplaceLemma(oldID int64, name, statement string) (*model.Lemma, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin lemma replacement: %w", err)
	}
	var draftID int64
	if err := tx.QueryRow(`SELECT draft_id FROM lemmas WHERE id=?`, oldID).Scan(&draftID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE lemmas SET status=? WHERE id=?`, string(model.LemmaReplaced), oldID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	res, err := tx.Exec(
		`INSERT INTO lemmas(draft_id,name,statement,status,created_at) VALUES(?,?,?,?,?)`,
		draftID, name, statement, string(model.LemmaAvailable), nowStr())
	if err != nil {
		_ = tx.Rollback()
		return nil, fmtErrInsertLemma(err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE premise_edges SET from_id=? WHERE draft_id=? AND from_kind='lemma' AND from_id=?`, newID, draftID, oldID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit lemma replacement: %w", err)
	}
	return s.GetLemma(newID)
}

func fmtErrInsertLemma(err error) error {
	return fmt.Errorf("insert lemma: %w", err)
}
