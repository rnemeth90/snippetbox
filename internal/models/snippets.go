package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (Snippet, error)
	Latest() ([]Snippet, error)
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB      *sql.DB
	Dialect string
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	statement := bindQuery(`SELECT id, title, content, created, expires FROM snippets
								WHERE expires > CURRENT_TIMESTAMP and id = ?`, m.Dialect)

	row := m.DB.QueryRow(statement, id)

	var s Snippet

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	statement := bindQuery(`INSERT INTO snippets (title, content, created, expires) VALUES(?,?,?,?)`, m.Dialect)
	now := time.Now().UTC()
	expiry := now.AddDate(0, 0, expires)

	var id int
	err := m.DB.QueryRow(statement+` RETURNING id`, title, content, now, expiry).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	statement := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > CURRENT_TIMESTAMP ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []Snippet
	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
