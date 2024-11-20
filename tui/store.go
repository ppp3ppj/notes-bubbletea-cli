package tui

import (
	"database/sql"
	"time"
  _ "github.com/mattn/go-sqlite3" // unknown driver sqlite3 forgotten import
)

type Note struct {
    Id int64
    Title string
    Body string
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
        Id interger not null primary key,
        Title text not null,
        Body text not null
    );`

    if _, err = s.conn.Exec(createTableStmt); err != nil {
        return err
    }

    return nil
}

func (s *Store) GetNotes() ([]Note, error) {
    rows, err := s.conn.Query("SELECT * FROM Notes")
    if err != nil {
        return nil, err
}

    defer rows.Close()
    notes := []Note{}
    for rows.Next() {
        var note Note
        rows.Scan(&note.Id, &note.Title, &note.Body)
        notes = append(notes, note)
    }

    return notes, nil
}

func (s *Store) SaveNote(note Note) error {
    if note.Id == 0 {
        note.Id = time.Now().UTC().Unix()
    }

    upsertQuery := `INSERT INTO Notes (Id, Title, Body)
    VALUES (?, ?, ?)
    ON CONFLICT(Id) DO UPDATE
    SET Title=excluded.Title, body=excluded.Body;
    `

    if _, err := s.conn.Exec(upsertQuery, note.Id, note.Title, note.Body); err != nil {
        return err
    }

    return nil
}
