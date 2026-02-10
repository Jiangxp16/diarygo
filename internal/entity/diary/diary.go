package diary

import (
	"diarygo/internal/db"
	"diarygo/internal/utils"
)

type Diary struct {
	ID       int    `json:"id"`
	Content  string `json:"content"`
	Weather  string `json:"weather"`
	Location string `json:"location"`
}

func (d *Diary) IsEmpty() bool {
	return d.Content == "" && d.Weather == "" && d.Location == ""
}

const TABLE = "diary"
const SQLCreate = `
	CREATE TABLE IF NOT EXISTS diary (
		id INTEGER PRIMARY KEY,
		content TEXT NOT NULL,
		weather CHAR(50) NOT NULL,
		location CHAR(50) NOT NULL
	);`

type Repository struct {
	*db.BaseRepository[Diary]
}

func NewRepository(d *db.DB) *Repository {
	base := db.NewBaseRepository[Diary](d, TABLE, SQLCreate, "")
	return &Repository{BaseRepository: base}
}

func (r *Repository) GetMonthDiary(year, month int) ([]*Diary, error) {
	d1 := utils.YMD2Int(year, month, 0)
	d2 := d1 + 31
	return r.GetBetween("id", d1, d2, "ASC")
}

func (r *Repository) Upsert(d *Diary) error {
	if d.IsEmpty() {
		if d.ID != 0 {
			return r.DeleteByID(d.ID)
		}
		return nil
	}
	return r.BaseRepository.Upsert(d)
}

func (r *Repository) UpsertMany(list []*Diary) error {
	if len(list) == 0 {
		return nil
	}
	toUpdate := []*Diary{}
	toDelete := []*Diary{}

	for _, d := range list {
		if d.IsEmpty() {
			toDelete = append(toDelete, d)
		} else {
			toUpdate = append(toUpdate, d)
		}
	}
	if len(toDelete) > 0 {
		if err := r.DeleteMany(toDelete); err != nil {
			return err
		}
	}
	if len(toUpdate) > 0 {
		if err := r.BaseRepository.UpsertMany(toUpdate); err != nil {
			return err
		}
	}
	return nil
}
