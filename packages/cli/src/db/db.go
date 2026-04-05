package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	ID        int64
	Content   string
	Source    string
	CreatedAt time.Time
	Pinned    bool
}

type DB struct {
	conn *sql.DB
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".clipboard", "clipboard.db")
}

func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS clips (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			content    TEXT NOT NULL,
			source     TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			pinned     INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clips(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_clips_pinned ON clips(pinned DESC, created_at DESC);
	`)
	return err
}

func (d *DB) Close() error { return d.conn.Close() }

// Add inserts a new clip. Returns error if duplicate within last 100.
func (d *DB) Add(content, source string) (*Entry, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("empty content")
	}

	// Deduplicate: check if exact content already exists recently
	var existing int64
	err := d.conn.QueryRow(`SELECT id FROM clips WHERE content = ? ORDER BY created_at DESC LIMIT 1`, content).Scan(&existing)
	if err == nil {
		// Already exists — update timestamp and return
		_, err = d.conn.Exec(`UPDATE clips SET created_at = CURRENT_TIMESTAMP WHERE id = ?`, existing)
		if err != nil {
			return nil, err
		}
		return d.Get(existing)
	}

	res, err := d.conn.Exec(
		`INSERT INTO clips (content, source) VALUES (?, ?)`,
		content, source,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.Get(id)
}

func (d *DB) Get(id int64) (*Entry, error) {
	row := d.conn.QueryRow(`SELECT id, content, source, created_at, pinned FROM clips WHERE id = ?`, id)
	return scanEntry(row)
}

func (d *DB) List(limit int, search string) ([]*Entry, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if search != "" {
		pattern := "%" + search + "%"
		rows, err = d.conn.Query(
			`SELECT id, content, source, created_at, pinned FROM clips
			 WHERE content LIKE ?
			 ORDER BY pinned DESC, created_at DESC LIMIT ?`,
			pattern, limit,
		)
	} else {
		rows, err = d.conn.Query(
			`SELECT id, content, source, created_at, pinned FROM clips
			 ORDER BY pinned DESC, created_at DESC LIMIT ?`,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (d *DB) Delete(id int64) error {
	res, err := d.conn.Exec(`DELETE FROM clips WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("clip %d not found", id)
	}
	return nil
}

func (d *DB) Pin(id int64, pin bool) error {
	val := 0
	if pin {
		val = 1
	}
	res, err := d.conn.Exec(`UPDATE clips SET pinned = ? WHERE id = ?`, val, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("clip %d not found", id)
	}
	return nil
}

func (d *DB) Clear(keepPinned bool) (int64, error) {
	var res sql.Result
	var err error
	if keepPinned {
		res, err = d.conn.Exec(`DELETE FROM clips WHERE pinned = 0`)
	} else {
		res, err = d.conn.Exec(`DELETE FROM clips`)
	}
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (d *DB) Stats() (total, pinned int64, err error) {
	d.conn.QueryRow(`SELECT COUNT(*) FROM clips`).Scan(&total)
	d.conn.QueryRow(`SELECT COUNT(*) FROM clips WHERE pinned = 1`).Scan(&pinned)
	return
}

// scanner interface covers both *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...any) error
}

func scanEntry(s scanner) (*Entry, error) {
	var e Entry
	var pinnedInt int
	err := s.Scan(&e.ID, &e.Content, &e.Source, &e.CreatedAt, &pinnedInt)
	if err != nil {
		return nil, err
	}
	e.Pinned = pinnedInt == 1
	return &e, nil
}
