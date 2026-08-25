package store

import (
	"fmt"

	"task239-lemmareview/internal/model"
)

// CreateEdge 新增前提依赖边（幂等：唯一索引冲突时忽略）。
func (s *Store) CreateEdge(e *model.PremiseEdge) error {
	t := nowStr()
	if e.FromKind == "" {
		e.FromKind = "lemma"
	}
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO premise_edges(draft_id,from_kind,from_id,to_step_id,required,created_at) VALUES(?,?,?,?,?,?)`,
		e.DraftID, e.FromKind, e.FromID, e.ToStepID, boolToInt(e.Required), t)
	if err != nil {
		return fmtErrInsertEdge(err)
	}
	return nil
}

// ListEdges 列出草稿内全部前提边。
func (s *Store) ListEdges(draftID int64) ([]*model.PremiseEdge, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,from_kind,from_id,to_step_id,required,created_at FROM premise_edges WHERE draft_id=? ORDER BY id`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.PremiseEdge
	for rows.Next() {
		e := &model.PremiseEdge{}
		var req int
		var ca string
		if err := rows.Scan(&e.ID, &e.DraftID, &e.FromKind, &e.FromID, &e.ToStepID, &req, &ca); err != nil {
			return nil, err
		}
		e.Required = true
		e.CreatedAt = parseTime(ca)
		out = append(out, e)
	}
	return out, nil
}

// ListEdgesForStep 返回某草稿中指向指定步骤的前提边，保持边的创建顺序。
func (s *Store) ListEdgesForStep(draftID, stepID int64) ([]*model.PremiseEdge, error) {
	rows, err := s.db.Query(`SELECT id,draft_id,from_kind,from_id,to_step_id,required,created_at FROM premise_edges WHERE draft_id=? AND to_step_id=? ORDER BY id`, draftID, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.PremiseEdge
	for rows.Next() {
		e := &model.PremiseEdge{}
		var req int
		var ca string
		if err := rows.Scan(&e.ID, &e.DraftID, &e.FromKind, &e.FromID, &e.ToStepID, &req, &ca); err != nil {
			return nil, err
		}
		e.Required = req != 0
		e.CreatedAt = parseTime(ca)
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteEdgesOfStep 删除某步骤的全部前提边（替换引理重算前清理）。
func (s *Store) DeleteEdgesOfStep(stepID int64) error {
	_, err := s.db.Exec(`DELETE FROM premise_edges WHERE to_step_id=?`, stepID)
	return err
}

func fmtErrInsertEdge(err error) error {
	return fmt.Errorf("insert edge: %w", err)
}
