package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/timofurrer/influss/internal/clip"
)

var (
	_ Store = (*SqlStore)(nil)
)

type SqlStore struct {
	db *sql.DB
}

func NewSqlStore(log *slog.Logger, connectionString string) (*SqlStore, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("sql connection string must be valid URL: %w", err)
	}

	var driver, dsn string
	switch u.Scheme {
	case "sqlite3":
		driver = "sqlite3"
		dsn = u.Host
	case "postgres":
		driver = "postgres"
		dsn = connectionString
	default:
		return nil, fmt.Errorf("sql connection string must either have sqlite3:// or postgres:// scheme")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql database: %w", err)
	}

	migrator := newMigrator(log, db)
	if err := migrator.run(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SqlStore{db: db}, nil
}

func (s *SqlStore) CreatedAt() time.Time {
	var createdAt sql.NullString
	err := s.db.QueryRow(`
		SELECT applied_at
		FROM schema_migration
		ORDER BY applied_at ASC
		LIMIT 1
	`).Scan(&createdAt)

	if err != nil {
		return time.Time{}
	}

	return parseTime(createdAt)
}

func (s *SqlStore) Load(ctx context.Context, lastN int) []*clip.Clip {
	query := `
		SELECT
			url, title, author,
			published_at, modified_at,
			excerpt, html_content
		FROM clip
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := s.db.QueryContext(ctx, query, lastN)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var clips []*clip.Clip
	for rows.Next() {
		c := &clip.Clip{}

		// Use temporary variables for timestamp fields
		var publishedAt, modifiedAt sql.NullString

		err := rows.Scan(
			&c.URL,
			&c.Title,
			&c.Author,
			&publishedAt,
			&modifiedAt,
			&c.Excerpt,
			&c.HTMLContent,
		)
		if err != nil {
			continue // Skip failed rows but continue processing
		}

		// Parse the timestamps
		c.PublishedAt = parseTime(publishedAt)
		c.ModifiedAt = parseTime(modifiedAt)

		clips = append(clips, c)
	}

	// If we hit an error during iteration, return what we have so far
	if rows.Err() != nil {
		return clips
	}

	return clips
}

func (s *SqlStore) Store(ctx context.Context, clip *clip.Clip) error {
	query := `
		INSERT INTO clip (
			url, title, author,
			published_at, modified_at,
			excerpt, html_content, plain_text_content
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			published_at = EXCLUDED.published_at,
			modified_at = EXCLUDED.modified_at,
			excerpt = EXCLUDED.excerpt,
			html_content = EXCLUDED.html_content,
			plain_text_content = EXCLUDED.plain_text_content,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		clip.URL,
		clip.Title,
		clip.Author,
		clip.PublishedAt,
		clip.ModifiedAt,
		clip.Excerpt,
		clip.HTMLContent,
		clip.PlainTextContent,
	)
	return err
}

func parseTime(s sql.NullString) time.Time {
	if s.Valid {
		t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", s.String)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
