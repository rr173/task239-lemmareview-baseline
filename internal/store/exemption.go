package store

import (
	"task239-lemmareview/internal/model"
)

// CreateExemption 新增假设豁免。
func (s *Store) CreateExemption(ex *model.Exemption) error {
	t := nowStr()
	_, err := s.db.Exec(
		`INSERT INTO exemptions(draft_id,step_id,lemma_id,reason,created_at) VALUES(?,?,?,?,?)`,
		ex.DraftID, ex.StepID, ex.LemmaID, ex.Reason, t)
	return err
}

// ListExemptions 列出草稿内豁免。
func (s *Store) ListExemptions(draftID int64) ([]*model.Exemption, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,step_id,lemma_id,reason,created_at FROM exemptions WHERE draft_id=? ORDER BY id`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Exemption
	for rows.Next() {
		ex := &model.Exemption{}
		var ca string
		if err := rows.Scan(&ex.ID, &ex.DraftID, &ex.StepID, &ex.LemmaID, &ex.Reason, &ca); err != nil {
			return nil, err
		}
		ex.CreatedAt = parseTime(ca)
		out = append(out, ex)
	}
	return out, nil
}
