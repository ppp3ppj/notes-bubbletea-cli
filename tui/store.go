package tui

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // unknown driver sqlite3 forgotten import
)

type Note struct {
	Id        string
	Title     string
	Body      string
	TotalTime string
    Project Project
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Project struct {
	Id          int
	Name        string
	Description string
    Categories []Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
    Id int
    Name string
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

	createTableProjectStmt := `
        CREATE TABLE IF NOT EXISTS Projects (
            Id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
            Name TEXT NOT NULL UNIQUE,
            Description TEXT,
            CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`

	createTableNoteStmt := `CREATE TABLE IF NOT EXISTS Notes (
        Id TEXT not null primary key,
        Title text not null,
        Body text not null,
        TotalTime TEXT,
        ProjectId INTEGER NOT NULL,
        CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (ProjectId) REFERENCES Projects(Id)
    );`

	if _, err = s.conn.Exec(createTableProjectStmt); err != nil {
		return err
	}

	if _, err = s.conn.Exec(createTableNoteStmt); err != nil {
		return err
	}


    // Insert mock projects if none exist
    mockProjects := []Project {
        {Name: "Work", Description: "Work-related tasks"},
        {Name: "Personal", Description: "Personal notes and ideas"},
        {Name: "Hobbies", Description: "Notes for hobbies and interests"},
    }

    for _, project := range mockProjects {
        if err := s.SaveProject(project); err != nil {
            // Ignore duplicate entries
            continue
        }
    }

	return nil
}

/*
func (s *Store) GetNotes() ([]Note, error) {
	rows, err := s.conn.Query("SELECT Id, Title, Body, TotalTime, CreatedAt, UpdatedAt FROM Notes")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var note Note
		rows.Scan(&note.Id, &note.Title, &note.Body, &note.TotalTime, &note.CreatedAt, &note.UpdatedAt)
		notes = append(notes, note)
	}

	return notes, nil
}
*/

func (s *Store) GetNotes() ([]Note, error) {
	query := `
		SELECT
			n.Id, n.Title, n.Body, n.TotalTime, n.CreatedAt, n.UpdatedAt,
			p.Id AS ProjectId, p.Name AS ProjectName, p.Description AS ProjectDescription
		FROM Notes n
		INNER JOIN Projects p ON n.ProjectId = p.Id;
	`

	rows, err := s.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		var project Project
		if err := rows.Scan(
			&note.Id, &note.Title, &note.Body, &note.TotalTime, &note.CreatedAt, &note.UpdatedAt,
			&project.Id, &project.Name, &project.Description,
		); err != nil {
			return nil, err
		}
		note.Project = project // Attach the project details to the note
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

	upsertQuery := `INSERT INTO Notes (Id, Title, Body, TotalTime, CreatedAt, UpdatedAt)
    VALUES (?, ?, ?, ?, ?, ?)
    ON CONFLICT(Id) DO UPDATE
    SET
        Title=excluded.Title,
        body=excluded.Body,
        TotalTime=excluded.TotalTime,
        UpdatedAt=excluded.UpdatedAt;
    `

	if _, err := s.conn.Exec(upsertQuery, note.Id, note.Title, note.Body, note.TotalTime, note.CreatedAt, note.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteNote(noteId string) error {
	_, err := s.conn.Exec("DELETE FROM notes WHERE id = ?", noteId)
	return err
}

func (s *Store) SaveProject(project Project) error {
	now := time.Now().UTC()

	insertQuery := `
    INSERT INTO Projects (Name, Description, CreatedAt, UpdatedAt)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(Name) DO NOTHING;
    `

	if _, err := s.conn.Exec(insertQuery, project.Name, project.Description, now, now); err != nil {
		return nil
	}

	return nil
}

func (s *Store) GetProjects() ([]Project, error) {
	rows, err := s.conn.Query("SELECT Id, Name, Description, CreatedAt, UpdatedAt FROM Projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []Project{}
	for rows.Next() {
		var project Project
		rows.Scan(&project.Id, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		projects = append(projects, project)
	}

	return projects, nil
}

func (s *Store) SaveNoteWithProject(note Note, projectId int) error {
	now := time.Now().UTC()

	if note.Id == "" {
		note.Id = uuid.New().String()
		note.CreatedAt = now
		note.UpdatedAt = now
	} else {
		note.UpdatedAt = now
	}

	upsertQuery := `INSERT INTO Notes (Id, Title, Body, TotalTime, ProjectId, CreatedAt, UpdatedAt)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(Id) DO UPDATE
    SET
        Title=excluded.Title,
        Body=excluded.Body,
        TotalTime=excluded.TotalTime,
        ProjectId=excluded.ProjectId,
        UpdatedAt=excluded.UpdatedAt;`

	if _, err := s.conn.Exec(upsertQuery, note.Id, note.Title, note.Body, note.TotalTime, projectId, note.CreatedAt, note.UpdatedAt); err != nil {
		return err
	}

	return nil
}


func (s *Store) GetNotesByProject(projectId int) ([]Note, error) {
	rows, err := s.conn.Query(
		"SELECT Id, Title, Body, TotalTime, CreatedAt, UpdatedAt FROM Notes WHERE ProjectId = ?", projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var note Note
		rows.Scan(&note.Id, &note.Title, &note.Body, &note.TotalTime, &note.CreatedAt, &note.UpdatedAt)
		notes = append(notes, note)
	}

	return notes, nil
}

func (s *Store) GetProjectById(projectId int) (Project, error) {
	var project Project
	query := "SELECT Id, Name, Description, CreatedAt, UpdatedAt FROM Projects WHERE Id = ?"
	err := s.conn.QueryRow(query, projectId).Scan(
		&project.Id, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Project{}, nil // Return zero value
	} else if err != nil {
		return Project{}, err
	}
	return project, nil
}
