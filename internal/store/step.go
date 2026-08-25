package store

import (
	"database/sql"
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateStep 新建推导步骤。
func (s *Store) CreateStep(step *model.Step) (*model.Step, error) {
	if step.DraftID == 0 {
		return nil, errDraftRequired
	}
	if step.Status == "" {
		step.Status = model.StepParsed
	}
	_, err := s.insertStep(s.db, step)
	if err != nil {
		return nil, fmtErrInsertStep(err)
	}
	return s.GetStep(step.ID)
}

// CreateSteps 原子写入一个步骤批次，任何一条失败都会回滚整批。
func (s *Store) CreateSteps(steps []*model.Step) ([]*model.Step, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin step batch: %w", err)
	}
	for _, step := range steps {
		if step.DraftID == 0 {
			_ = tx.Rollback()
			return nil, errDraftRequired
		}
		if step.Status == "" {
			step.Status = model.StepParsed
		}
		if _, err := s.insertStep(tx, step); err != nil {
			_ = tx.Rollback()
			return nil, fmtErrInsertStep(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit step batch: %w", err)
	}
	out := make([]*model.Step, 0, len(steps))
	for _, step := range steps {
		created, err := s.GetStep(step.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, created)
	}
	return out, nil
}

type stepInserter interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func (s *Store) insertStep(exec stepInserter, step *model.Step) (int64, error) {
	res, err := exec.Exec(
		`INSERT INTO steps(draft_id,seq,label,statement,conclusion,status,created_at) VALUES(?,?,?,?,?,?,?)`,
		step.DraftID, step.Seq, step.Label, step.Statement, step.Conclusion, string(step.Status), nowStr())
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	step.ID = id
	return id, nil
}

// GetStep 查询步骤。
func (s *Store) GetStep(id int64) (*model.Step, error) {
	row := s.db.QueryRow(`SELECT id,draft_id,seq,label,statement,conclusion,status,created_at FROM steps WHERE id=?`, id)
	st := &model.Step{}
	var status, ca string
	if err := row.Scan(&st.ID, &st.DraftID, &st.Seq, &st.Label, &st.Statement, &st.Conclusion, &status, &ca); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrStepNotFound
		}
		return nil, err
	}
	st.Status = model.StepStatus(status)
	st.CreatedAt = parseTime(ca)
	return st, nil
}

// ListSteps 列出草稿内步骤（按 seq 排序）。
func (s *Store) ListSteps(draftID int64) ([]*model.Step, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,seq,label,statement,conclusion,status,created_at FROM steps WHERE draft_id=? ORDER BY seq`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Step
	for rows.Next() {
		st := &model.Step{}
		var status, ca string
		if err := rows.Scan(&st.ID, &st.DraftID, &st.Seq, &st.Label, &st.Statement, &st.Conclusion, &status, &ca); err != nil {
			return nil, err
		}
		st.Status = model.StepStatus(status)
		st.CreatedAt = parseTime(ca)
		out = append(out, st)
	}
	return out, nil
}

// UpdateStepConclusion 更新步骤结论。
func (s *Store) UpdateStepConclusion(id int64, conclusion string) error {
	res, err := s.db.Exec(`UPDATE steps SET conclusion=?, status=? WHERE id=?`, conclusion, string(model.StepParsed), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrStepNotFound
	}
	return nil
}

// SetStepStatus 设置步骤状态。
func (s *Store) SetStepStatus(id int64, status model.StepStatus) error {
	res, err := s.db.Exec(`UPDATE steps SET status=? WHERE id=?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrStepNotFound
	}
	return nil
}
