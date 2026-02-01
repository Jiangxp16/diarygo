package interest

import (
	"diarygo/internal/db"
	"diarygo/internal/utils"
)

// Interest model
type Interest struct {
	ID         int     `json:"id"`
	Added      int     `json:"added"`
	Updated    int     `json:"updated"`
	Name       string  `json:"name"`
	Sort       int     `json:"sort"`
	Progress   string  `json:"progress"`
	Publish    int     `json:"publish"`
	Date       int     `json:"date"`
	Score_DB   float64 `json:"score_db"`   //lint:ignore ST1003
	Score_IMDB float64 `json:"score_imdb"` //lint:ignore ST1003
	Score      float64 `json:"score"`
	Remark     string  `json:"remark"`
}

func (i *Interest) IsEmpty() bool {
	return i.Name == "" && i.Remark == ""
}

// -------------------- Repository --------------------

const TABLE = "interest"
const SQLCreate = `
	CREATE TABLE IF NOT EXISTS interest (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		added INTEGER NOT NULL DEFAULT 0,
		updated INTEGER NOT NULL DEFAULT 0,
		name CHAR(255) NOT NULL DEFAULT '',
		sort INTEGER NOT NULL DEFAULT 0,
		progress CHAR(20) NOT NULL DEFAULT '',
		publish INTEGER NOT NULL DEFAULT 0,
		date INTEGER NOT NULL DEFAULT 0,
		score_db REAL NOT NULL DEFAULT -1,
		score_imdb REAL NOT NULL DEFAULT -1,
		score REAL NOT NULL DEFAULT -1,
		remark TEXT NOT NULL
	);
	`
const SQLExtra = `UPDATE interest SET sort=7 WHERE sort=0;`

type Repository struct {
	*db.BaseRepository[Interest]
}

func NewRepository(d *db.DB) *Repository {
	base := db.NewBaseRepository[Interest](d, TABLE, SQLCreate, SQLExtra)
	return &Repository{BaseRepository: base}
}

func (r *Repository) GetBySort(sort int) ([]*Interest, error) {
	if sort == 0 {
		return r.GetList("ORDER BY date DESC")
	}
	return r.GetList("WHERE sort=? ORDER BY date DESC", sort)
}

func (r *Repository) Add(i *Interest) (*Interest, error) {
	if i.Sort == 0 {
		i.Sort = 7
	}
	if i.Added == 0 {
		i.Added = utils.GetCurrentDateInt()
	}
	if i.Date == 0 {
		i.Date = utils.GetCurrentYYYYMM()
	}
	if i.Score_DB == 0 {
		i.Score_DB = -1
	}
	if i.Score_IMDB == 0 {
		i.Score_IMDB = -1
	}
	if i.Score == 0 {
		i.Score = -1
	}
	return r.BaseRepository.Add(i)
}
