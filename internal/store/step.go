package store

import (
	"database/sql"

	"task239-lemmareview/internal/model"
)

// CreateStep 新建推导步骤。
func (s *Store) CreateStep(step *model.Step) (*model.Step, error) {
	if step.DraftID == 0 {
		return nil, errDraftRequired
	}
	t := nowStr()
	if step.Status == "" {
		step.Status = model.StepParsed
	}
	res, err := s.db.Exec(
		`INSERT INTO steps(draft_id,seq,label,statement,conclusion,status,created_at) VALUES(?,?,?,?,?,?,?)`,
		step.DraftID, step.Seq, step.Label, step.Statement, step.Conclusion, string(step.Status), t)
	if err != nil {
		return nil, fmtErrInsertStep(err)
	}
	id, _ := res.LastInsertId()
	return s.GetStep(id)
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
