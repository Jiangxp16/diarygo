package db

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type BaseRepository[T any] struct {
	DB        *DB
	Table     string
	SQLCreate string
	SQLExtra  string
}

func NewBaseRepository[T any](d *DB, table string, createSQL, extraSQL string) *BaseRepository[T] {
	if createSQL != "" {
		d.Exec(createSQL, nil, false)
	}
	if extraSQL != "" {
		d.Exec(extraSQL, nil, false)
	}
	return &BaseRepository[T]{
		DB:        d,
		Table:     table,
		SQLCreate: createSQL,
		SQLExtra:  extraSQL,
	}
}

func (r *BaseRepository[T]) Add(v *T) (*T, error) {
	return Add(r.DB, r.Table, v)
}

func (r *BaseRepository[T]) AddMany(list []*T) error {
	return AddMany(r.DB, r.Table, list)
}

func (r *BaseRepository[T]) DeleteWhere(query string, args ...any) error {
	return DeleteWhere(r.DB, r.Table, query, args...)
}

func (r *BaseRepository[T]) DeleteByID(id int) error {
	return DeleteByID(r.DB, r.Table, id)
}

func (r *BaseRepository[T]) Delete(v *T) error {
	return Delete(r.DB, r.Table, v)
}

func (r *BaseRepository[T]) DeleteMany(list []*T) error {
	return DeleteMany(r.DB, r.Table, list)
}

func (r *BaseRepository[T]) Update(v *T) error {
	return Update(r.DB, r.Table, v)
}

func (r *BaseRepository[T]) UpdateByID(id int, params map[string]any) error {
	return UpdateByID(r.DB, r.Table, id, params)
}

func (r *BaseRepository[T]) UpdateMany(list []*T) error {
	return UpdateMany(r.DB, r.Table, list)
}

func (r *BaseRepository[T]) Upsert(v *T) error {
	return Upsert(r.DB, r.Table, v)
}

func (r *BaseRepository[T]) UpsertMany(list []*T) error {
	return UpdateMany(r.DB, r.Table, list)
}

func (r *BaseRepository[T]) List() ([]*T, error) {
	return SelectList[T](r.DB, r.Table, "")
}

func (r *BaseRepository[T]) GetLast() (*T, error) {
	return SelectLast[T](r.DB, r.Table)
}

func (r *BaseRepository[T]) GetByID(id int) (*T, error) {
	return SelectByID[T](r.DB, r.Table, id)
}

func (r *BaseRepository[T]) GetList(query string, args ...any) ([]*T, error) {
	return SelectList[T](r.DB, r.Table, query, args...)
}

func (r *BaseRepository[T]) GetBetween(field string, d1, d2 int, order string) ([]*T, error) {
	return SelectBetween[T](r.DB, r.Table, field, d1, d2, fmt.Sprintf("%s %s", field, order))
}

func (r *BaseRepository[T]) GetBetweenDates(d1, d2 int, order string) ([]*T, error) {
	return r.GetBetween("date", d1, d2, order)
}

func (r *BaseRepository[T]) Reset() error {
	_, err := r.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", r.Table), nil, true)
	if err != nil {
		return err
	}
	if r.SQLCreate != "" {
		_, err = r.DB.Exec(r.SQLCreate, nil, false)
		if err != nil {
			return err
		}
	}
	if r.SQLExtra != "" {
		_, err = r.DB.Exec(r.SQLExtra, nil, false)
		if err != nil {
			return err
		}
	}
	return err
}

func (r *BaseRepository[T]) Export(w io.Writer) error {
	list, err := r.GetList("")
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return errors.New("empty list")
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := StructCols(list[0], false)
	if err := f.SetSheetRow(sheet, "A1", &headers); err != nil {
		return err
	}

	for i, item := range list {
		row := i + 2
		data := StructArgs(item, false)
		if err := f.SetSheetRow(sheet, "A"+strconv.Itoa(row), &data); err != nil {
			return err
		}
	}

	return f.Write(w)
}

func FillStructFromRow(v any, row []string) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("v must be pointer to struct")
	}
	val = val.Elem()

	numField := val.NumField()

	setField := func(field reflect.Value, s string) {
		if s == "" {
			return
		}
		switch field.Kind() {
		case reflect.Int, reflect.Int64:
			i, _ := strconv.ParseInt(s, 10, 64)
			field.SetInt(i)
		case reflect.Float64:
			f, _ := strconv.ParseFloat(s, 64)
			field.SetFloat(f)
		case reflect.String:
			field.SetString(s)
		}
	}
	for i := 0; i < numField && i < len(row); i++ {
		setField(val.Field(i), row[i])
	}
	return nil
}

func (r *BaseRepository[T]) Import(reader io.Reader) error {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}
	var toUpsert []*T
	for i, row := range rows {
		if i == 0 {
			continue
		}
		item := new(T)
		err := FillStructFromRow(item, row)
		if err != nil {
			return err
		}
		toUpsert = append(toUpsert, item)
	}
	if len(toUpsert) > 0 {
		if err := UpsertMany(r.DB, r.Table, toUpsert); err != nil {
			return err
		}
	}
	return nil
}
