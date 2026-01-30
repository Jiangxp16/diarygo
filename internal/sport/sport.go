package sport

import (
	"diarygo/internal/db"
	"diarygo/internal/utils"
)

type Sport struct {
	ID      int    `json:"id"`
	Date    int    `json:"date"`
	Content string `json:"content"`
}

const TABLE = "sport"
const SQLCreate = `
	CREATE TABLE IF NOT EXISTS sport (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date INTEGER NOT NULL DEFAULT 0,
		content TEXT NOT NULL DEFAULT ""
	);`

type Repository struct {
	*db.BaseRepository[Sport]
}

func NewRepository(d *db.DB) *Repository {
	base := db.NewBaseRepository[Sport](d, TABLE, SQLCreate, "")
	return &Repository{BaseRepository: base}
}

func (r *Repository) Add(n *Sport) (*Sport, error) {
	if n.Date == 0 {
		n.Date = utils.GetCurrentDateInt()
	}
	return r.BaseRepository.Add(n)
}
