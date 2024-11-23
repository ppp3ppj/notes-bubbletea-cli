package tui

import (
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // unknown driver sqlite3 forgotten import
	"time"
)

type Note struct {
	Id        string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store struct {
	conn *sql.DB
}

func (s *Store) Init() error {
	var err error
	s.conn, err = sql.Open("sqlite3", "./notes.db")
	if err != nil {
		return err
	}

	createTableStmt := `CREATE TABLE IF NOT EXISTS Notes (
        Id TEXT not null primary key,
        Title text not null,
        Body text not null,
        CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	if _, err = s.conn.Exec(createTableStmt); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetNotes() ([]Note, error) {
	rows, err := s.conn.Query("SELECT Id, Title, Body, CreatedAt, UpdatedAt FROM Notes")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var note Note
		rows.Scan(&note.Id, &note.Title, &note.Body, &note.CreatedAt, &note.UpdatedAt)
		notes = append(notes, note)
	}

	return notes, nil
}

func (s *Store) SaveNote(note Note) error {
	now := time.Now().UTC()

	if note.Id == "" {
		note.Id = uuid.New().String()
		note.CreatedAt = now
		note.UpdatedAt = now
	} else {
		note.UpdatedAt = now
	}

	upsertQuery := `INSERT INTO Notes (Id, Title, Body, CreatedAt, UpdatedAt)
    VALUES (?, ?, ?, ?, ?)
    ON CONFLICT(Id) DO UPDATE
    SET
        Title=excluded.Title,
        body=excluded.Body,
        UpdatedAt=excluded.UpdatedAt;
    `

	if _, err := s.conn.Exec(upsertQuery, note.Id, note.Title, note.Body, note.CreatedAt, note.UpdatedAt); err != nil {
		return err
	}

	return nil
}
