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
	Project   Project
	Category  Category
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Project struct {
	Id          int
	Name        string
	Description string
	Categories  []Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	Id   int
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
        CategoryId INTEGER,
        CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (ProjectId) REFERENCES Projects(Id),
        FOREIGN KEY (CategoryId) REFERENCES Categories(Id)
    );`

	createTableCategoryStmt := `
        CREATE TABLE IF NOT EXISTS Categories (
            Id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
            Name TEXT NOT NULL UNIQUE
        );
    `

	createTableProjectCategoriesStmt := `
        CREATE TABLE IF NOT EXISTS ProjectCategories (
            ProjectId INTEGER NOT NULL,
            CategoryId INTEGER NOT NULL,
            PRIMARY KEY (ProjectId, CategoryId),
            FOREIGN KEY (ProjectId) REFERENCES Projects(Id),
            FOREIGN KEY (CategoryId) REFERENCES Categories(Id)
        );
    `

	if _, err = s.conn.Exec(createTableProjectStmt); err != nil {
		return err
	}

	if _, err = s.conn.Exec(createTableNoteStmt); err != nil {
		return err
	}

	if _, err = s.conn.Exec(createTableCategoryStmt); err != nil {
		return err
	}

	if _, err = s.conn.Exec(createTableProjectCategoriesStmt); err != nil {
		return err
	}

	// Insert mock projects if none exist
	mockProjects := []Project{
		{Name: "Work", Description: "Work-related tasks"},
		{Name: "Personal", Description: "Personal notes and ideas"},
		{Name: "Hobbies", Description: "Notes for hobbies and interests"},
		{Name: "General", Description: "Notes for general idea"},
	}

	for _, project := range mockProjects {
		if err := s.SaveProject(project); err != nil {
			// Ignore duplicate entries
			continue
		}
	}

	// Insert mock categories and project categories
	mockCategories := []Category{
		{Name: "Urgent"},
		{Name: "Important"},
		{Name: "Optional"},
	}

	for _, category := range mockCategories {
		query := `INSERT OR IGNORE INTO Categories (Name) VALUES (?);`
		if _, err := s.conn.Exec(query, category.Name); err != nil {
			return err
		}
	}

	// Link categories to projects (mock)
	mockAssignments := map[string][]string{
		"Work":     {"Urgent", "Important"},
		"Personal": {"Important", "Optional"},
		"Hobbies":  {"Optional"},
	}

	for projectName, categoryNames := range mockAssignments {
		project, err := s.GetProjectByName(projectName)
		if err != nil || project.Id == 0 {
			continue
		}

		for _, categoryName := range categoryNames {
			var categoryId int
			err := s.conn.QueryRow(`SELECT Id FROM Categories WHERE Name = ?`, categoryName).Scan(&categoryId)
			if err != nil {
				continue
			}

			// Assign categories to the project
			if err := s.AssignCategoriesToProject(project.Id, []int{categoryId}); err != nil {
				return err
			}
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
			p.Id AS ProjectId, p.Name AS ProjectName, p.Description AS ProjectDescription,
			c.Id AS CategoryId, c.Name AS CategoryName
		FROM Notes n
		INNER JOIN Projects p ON n.ProjectId = p.Id
		LEFT JOIN Categories c ON n.CategoryId = c.Id;
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
		var category Category
		if err := rows.Scan(
			&note.Id, &note.Title, &note.Body, &note.TotalTime, &note.CreatedAt, &note.UpdatedAt,
			&project.Id, &project.Name, &project.Description,
			&category.Id, &category.Name,
		); err != nil {
			return nil, err
		}
		note.Project = project // Attach the project details to the note
		note.Category = category // Attach the category detail to the note
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

func (s *Store) SaveNoteWithProject(note Note, projectId, category int, currentdate time.Time) error {
	now := time.Now().UTC()

	if note.Id == "" {
		note.Id = uuid.New().String()
		note.CreatedAt = currentdate.UTC()
		note.UpdatedAt = currentdate.UTC()
	} else {
		note.UpdatedAt = now
	}

	upsertQuery := `INSERT INTO Notes (Id, Title, Body, TotalTime, ProjectId, CategoryId, CreatedAt, UpdatedAt)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(Id) DO UPDATE
    SET
        Title=excluded.Title,
        Body=excluded.Body,
        TotalTime=excluded.TotalTime,
        ProjectId=excluded.ProjectId,
        CategoryId=excluded.CategoryId,
        UpdatedAt=excluded.UpdatedAt;`

	if _, err := s.conn.Exec(upsertQuery, note.Id, note.Title, note.Body, note.TotalTime, projectId, category, note.CreatedAt, note.UpdatedAt); err != nil {
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

func (s *Store) AssignCategoriesToProject(projectId int, categoryIds []int) error {
	query := "INSERT OR IGNORE INTO ProjectCategories (ProjectId, CategoryId) VALUES (?, ?)"
	for _, categoryId := range categoryIds {
		if _, err := s.conn.Exec(query, projectId, categoryId); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetCategoriesByProject(projectId int) ([]Category, error) {
	query := `
        SELECT c.Id, c.Name
        FROM Categories c
        INNER JOIN ProjectCategories pc ON c.Id = pc.CategoryId
        WHERE pc.ProjectId = ?
    `
	rows, err := s.conn.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.Id, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (s *Store) GetProjectByName(name string) (Project, error) {
	var project Project
	query := `SELECT Id, Name, Description, CreatedAt, UpdatedAt FROM Projects WHERE Name = ?`
	err := s.conn.QueryRow(query, name).Scan(
		&project.Id, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Project{}, nil // Return zero value
	} else if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) UpdateNoteCategory(noteId string, categoryId int) error {
	query := `
		UPDATE Notes
		SET CategoryId = ?
		WHERE Id = ?;
	`
	_, err := s.conn.Exec(query, categoryId, noteId)
	return err
}

func (s *Store) GetNotesByDate(currentDate time.Time) ([]Note, error) {
	query := `
        SELECT
			n.Id, n.Title, n.Body, n.TotalTime, n.CreatedAt, n.UpdatedAt,
			p.Id AS ProjectId, p.Name AS ProjectName, p.Description AS ProjectDescription,
			c.Id AS CategoryId, c.Name AS CategoryName
		FROM Notes n
		INNER JOIN Projects p ON n.ProjectId = p.Id
		LEFT JOIN Categories c ON n.CategoryId = c.Id
		WHERE date(n.CreatedAt) = date(?);
	`

	rows, err := s.conn.Query(query, currentDate.UTC().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		var project Project
		var category Category
		if err := rows.Scan(
			&note.Id, &note.Title, &note.Body, &note.TotalTime, &note.CreatedAt, &note.UpdatedAt,
			&project.Id, &project.Name, &project.Description,
			&category.Id, &category.Name,
		); err != nil {
			return nil, err
		}
		note.Project = project // Attach the project details to the note
		note.Category = category // Attach the category details to the note
		notes = append(notes, note)
	}
	return notes, nil
}
