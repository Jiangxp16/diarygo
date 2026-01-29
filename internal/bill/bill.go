package bill

import (
	"diarygo/internal/db"
	"diarygo/internal/utils"
)

// Bill model
type Bill struct {
	ID     int     `json:"id"`
	Date   int     `json:"date"`
	Inout  int     `json:"inout"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	Item   string  `json:"item"`
}

const TABLE = "bill"
const SQLCreate = `
	CREATE TABLE IF NOT EXISTS bill (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date INTEGER NOT NULL DEFAULT 0,
		inout INTEGER NOT NULL DEFAULT -1,
		type CHAR(20) NOT NULL DEFAULT "",
		amount DECIMAL(10,2) NOT NULL DEFAULT 0.0,
		item TEXT NOT NULL DEFAULT ""
	);`

type Repository struct {
	*db.BaseRepository[Bill]
}

func NewRepository(d *db.DB) *Repository {
	base := db.NewBaseRepository[Bill](d, TABLE, SQLCreate, "")
	return &Repository{BaseRepository: base}
}

func (r *Repository) Add(b *Bill) (*Bill, error) {
	if b.Date == 0 {
		b.Date = utils.GetCurrentDateInt()
	}
	if b.Inout == 0 {
		b.Inout = -1
	}
	return r.BaseRepository.Add(b)
}
