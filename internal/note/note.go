package note

import (
	"diarygo/internal/db"
	"diarygo/internal/utils"
)

type Note struct {
	ID       int    `json:"id"`
	Begin    int    `json:"begin"`
	Last     int    `json:"last"`
	Process  int    `json:"process"`
	Desire   int    `json:"desire"`
	Priority int    `json:"priority"`
	Content  string `json:"content"`
}

const TABLE = "note"
const SQLCreate = `
	CREATE TABLE IF NOT EXISTS note (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		begin INTEGER NOT NULL DEFAULT 0,
		last INTEGER NOT NULL DEFAULT 0,
		process INTEGER NOT NULL DEFAULT 0,
		desire INTEGER NOT NULL DEFAULT 0,
		priority INTEGER NOT NULL DEFAULT 0,
		content TEXT NOT NULL DEFAULT ""
	);`

type Repository struct {
	*db.BaseRepository[Note]
}

func NewRepository(d *db.DB) *Repository {
	base := db.NewBaseRepository[Note](d, TABLE, SQLCreate, "")
	return &Repository{BaseRepository: base}
}

func (r *Repository) Add(n *Note) (*Note, error) {
	if n.Begin == 0 {
		n.Begin = utils.GetCurrentDateInt()
	}
	if n.Last == 0 {
		n.Last = utils.GetCurrentDateInt()
	}
	return r.BaseRepository.Add(n)
}

func (r *Repository) List() ([]*Note, error) {
	return r.GetList("ORDER BY last DESC")
}
