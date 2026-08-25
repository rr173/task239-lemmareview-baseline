// Package store 负责 SQLite 持久化：建表迁移与所有实体的 CRUD。
// 使用纯 Go 驱动 modernc.org/sqlite，CGO 无关，离线可构建。
// 各实体的 CRUD 按文件拆分为 draft.go / step.go / lemma.go / edge.go /
// exemption.go / version.go，本文件仅含连接管理与建表迁移。
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 连接与所有持久化操作。
type Store struct {
	db *sql.DB
}

// New 打开 SQLite 数据库并执行建表迁移。
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写者，避免并发写锁竞争
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set wal: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable fk: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return s.db.Close()
}

// DB 返回底层连接（供测试与事务使用）。
func (s *Store) DB() *sql.DB { return s.db }

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS drafts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS steps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL REFERENCES drafts(id),
			seq INTEGER NOT NULL,
			label TEXT NOT NULL,
			statement TEXT NOT NULL DEFAULT '',
			conclusion TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_step_draft_label ON steps(draft_id, label);`,
		`CREATE TABLE IF NOT EXISTS lemmas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL REFERENCES drafts(id),
			name TEXT NOT NULL,
			statement TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS premise_edges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL REFERENCES drafts(id),
			from_kind TEXT NOT NULL,
			from_id INTEGER NOT NULL,
			to_step_id INTEGER NOT NULL REFERENCES steps(id),
			required INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_edge ON premise_edges(draft_id, from_kind, from_id, to_step_id);`,
		`CREATE TABLE IF NOT EXISTS exemptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL REFERENCES drafts(id),
			step_id INTEGER NOT NULL REFERENCES steps(id),
			lemma_id INTEGER NOT NULL DEFAULT 0,
			reason TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS proof_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_id INTEGER NOT NULL REFERENCES drafts(id),
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			lemma_set TEXT NOT NULL DEFAULT '',
			snapshot TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);`,
	}
	for _, st := range stmts {
		if _, err := s.db.Exec(st); err != nil {
			return fmt.Errorf("migrate failed: %w", err)
		}
	}
	return nil
}

func nowStr() string { return time.Now().UTC().Format(time.RFC3339) }

// parseTime 将存储的 RFC3339 字符串解析回 time.Time（解析失败回退零值）。
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// errDraftRequired 步骤/引理缺少草稿归属时的错误。
var errDraftRequired = fmt.Errorf("draft_id required")

// fmtErrInsertStep 包装步骤插入错误，便于调用方识别。
func fmtErrInsertStep(err error) error {
	return fmt.Errorf("insert step: %w", err)
}
