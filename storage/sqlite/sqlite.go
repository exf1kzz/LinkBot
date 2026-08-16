package sqlite

import (
	"LinkBot/storage"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, p *storage.Page) error {
	q := `INSERT OR IGNORE INTO pages (url, user_name) VALUES (?, ?)`

	if _, err := s.db.ExecContext(ctx, q, p.URL, p.UserName); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	q := `SELECT url FROM pages WHERE user_name = ? ORDER BY RANDOM() LIMIT 1`

	var url string

	err := s.db.QueryRowContext(ctx, q, userName).Scan(&url)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick random page: %w", err)
	}

	return &storage.Page{
		URL:      url,
		UserName: userName,
	}, nil
}

func (s *Storage) List(ctx context.Context, userName string) ([]storage.Page, error) {
	q := `SELECT url FROM pages WHERE user_name = ? ORDER BY url`

	rows, err := s.db.QueryContext(ctx, q, userName)
	if err != nil {
		return nil, fmt.Errorf("can't list pages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	pages := make([]storage.Page, 0)

	for rows.Next() {
		var page storage.Page
		if err := rows.Scan(&page.URL); err != nil {
			return nil, fmt.Errorf("can't scan page: %w", err)
		}

		page.UserName = userName
		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("can't iterate pages: %w", err)
	}

	if len(pages) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	return pages, nil
}

func (s *Storage) Remove(ctx context.Context, page *storage.Page) error {
	q := `DELETE FROM pages WHERE url = ? AND user_name = ?`

	if _, err := s.db.ExecContext(ctx, q, page.URL, page.UserName); err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

func (s *Storage) Exists(ctx context.Context, page *storage.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE url = ? AND user_name = ?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, page.URL, page.UserName).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check page existence: %w", err)
	}

	return count > 0, nil
}

func (s *Storage) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT NOT NULL, user_name TEXT NOT NULL)`

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	q = `DELETE FROM pages WHERE rowid NOT IN (SELECT MIN(rowid) FROM pages GROUP BY url, user_name)`

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("can't remove duplicate pages: %w", err)
	}

	q = `CREATE UNIQUE INDEX IF NOT EXISTS idx_pages_url_user_name ON pages(url, user_name)`

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("can't create pages index: %w", err)
	}

	return nil
}
