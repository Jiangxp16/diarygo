package db

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"diarygo/internal/utils"

	_ "modernc.org/sqlite" // pure Go sqlite driver
)

var Key string
var GlobalWriteMutex sync.Mutex

// DB 包装
type DB struct {
	Conn *sql.DB
}

// Open 打开数据库
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{Conn: conn}, nil
}

// Close 关闭数据库
func (d *DB) Close() error {
	if d.Conn != nil {
		return d.Conn.Close()
	}
	return nil
}

var (
	instance *DB
	once     sync.Once
	initErr  error
)

func Init(dbPath string) {
	once.Do(func() {
		if dbPath == "" {
			dbPath = "diary.db"
		}
		instance, initErr = Open(dbPath)
		instance.Exec(`PRAGMA journal_mode = WAL`, nil, false)
		instance.Exec(`PRAGMA synchronous = NORMAL`, nil, false)
		instance.Exec(`PRAGMA busy_timeout = 5000`, nil, false)
		instance.Exec(`PRAGMA wal_autocheckpoint = 1000`, nil, false)
	})
	if initErr != nil {
		panic(initErr)
	}
}

func Get() *DB {
	if instance == nil {
		panic("db not initialized, call db.Init() first")
	}
	return instance
}

func Close() {
	if instance != nil {
		_ = instance.Close()
	}
}

// ==========================
// 加密 / 解密函数
// ==========================

const headEncrypt = "__en__"

func encryptString(text string) string {
	if Key == "" || strings.HasPrefix(text, headEncrypt) {
		return text
	}
	runes := []rune(text)
	keyRunes := []rune(Key)
	for i := range runes {
		runes[i] ^= keyRunes[i%len(keyRunes)]
	}
	return headEncrypt + string(runes)
}

func decryptString(text string) string {
	if Key == "" || !strings.HasPrefix(text, headEncrypt) {
		return text
	}
	text = text[len(headEncrypt):]
	runes := []rune(text)
	keyRunes := []rune(Key)
	for i := range runes {
		runes[i] ^= keyRunes[i%len(keyRunes)]
	}
	return string(runes)
}

func EncryptArgs(args []any) []any {
	for i, v := range args {
		switch val := v.(type) {
		case string:
			args[i] = encryptString(val)
		case []any:
			args[i] = EncryptArgs(val)
		}
	}
	return args
}

func DecryptArgs(args []any) []any {
	for i, v := range args {
		switch val := v.(type) {
		case string:
			args[i] = decryptString(val)
		case []any:
			args[i] = DecryptArgs(val)
		}
	}
	return args
}

// ==========================
// 查询包装
// ==========================

// Exec 执行单条 SQL
func (d *DB) Exec(sqlCmd string, args []any, encrypt bool) (sql.Result, error) {
	if encrypt {
		args = EncryptArgs(args)
	}
	return d.Conn.Exec(sqlCmd, args...)
}

// ExecMany 批量执行
func (d *DB) ExecMany(sqlCmd string, argsList [][]any, encrypt bool) error {
	tx, err := d.Conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(sqlCmd)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, args := range argsList {
		if encrypt {
			args = EncryptArgs(args)
		}
		_, err := stmt.Exec(args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func normalizeDBType(t string) string {
	t = strings.ToUpper(t)
	if i := strings.Index(t, "("); i != -1 {
		t = t[:i]
	}
	return strings.TrimSpace(t)
}

func NormalizeValue(value any, colType *sql.ColumnType) any {
	switch t := value.(type) {
	case int64:
		switch normalizeDBType(colType.DatabaseTypeName()) {
		case "DECIMAL", "NUMERIC", "REAL", "FLOAT", "DOUBLE":
			return float64(t)
		}
	case int:
		switch normalizeDBType(colType.DatabaseTypeName()) {
		case "DECIMAL", "NUMERIC", "REAL", "FLOAT", "DOUBLE":
			return float64(t)
		}
	}
	return value
}

// Select 查询多行
func (d *DB) Select(sqlCmd string, args []any, decrypt bool) ([][]any, error) {
	rows, err := d.Conn.Query(sqlCmd, EncryptArgs(args)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colTypes, _ := rows.ColumnTypes()
	var results [][]any
	for rows.Next() {
		values := make([]any, len(colTypes))
		scanArgs := make([]any, len(colTypes))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		if decrypt {
			values = DecryptArgs(values)
		}
		for i, v := range values {
			values[i] = NormalizeValue(v, colTypes[i])
		}
		results = append(results, values)
	}
	return results, nil
}

func Update(d *DB, table string, v any) error {
	cols, args := StructCols(v, true), StructArgs(v, true)
	set := make([]string, 0, len(cols)-1)
	for _, c := range cols[:len(cols)-1] {
		set = append(set, c+"=?")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id=?", table, strings.Join(set, ", "))
	_, err := d.Exec(sql, args, true)
	return err
}

func UpdateByID(d *DB, table string, id any, params map[string]any) error {
	if len(params) == 0 {
		return errors.New("empty params")
	}

	set := make([]string, 0, len(params))
	args := make([]any, 0, len(params)+1)

	for k, v := range params {
		set = append(set, fmt.Sprintf("%s=?", k))
		args = append(args, v)
	}

	// id 放最后
	args = append(args, id)

	sql := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id=?",
		table,
		strings.Join(set, ", "),
	)

	_, err := d.Exec(sql, args, true)
	return err
}

func UpdateMany[T any](d *DB, table string, list []*T) error {
	if len(list) == 0 {
		return errors.New("empty list")
	}
	cols, argsList := StructCols(list[0], true), StructArgsList(list, true)
	set := make([]string, 0, len(cols)-1)
	for _, c := range cols[:len(cols)-1] {
		set = append(set, c+"=?")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id=?", table, strings.Join(set, ", "))
	return d.ExecMany(sql, argsList, true)
}

func StructCols(v any, idLast bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	numField := val.NumField()

	cols := make([]string, 0, numField)

	if !idLast {
		for i := 0; i < numField; i++ {
			cols = append(cols, strings.ToLower(typ.Field(i).Name))
		}
		return cols
	}

	for i := 1; i < numField; i++ {
		cols = append(cols, strings.ToLower(typ.Field(i).Name))
	}
	cols = append(cols, "id")
	return cols
}

func StructArgs(v any, idLast bool) []any {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	numField := val.NumField()
	args := make([]any, 0, numField)

	if !idLast {
		for i := 0; i < numField; i++ {
			args = append(args, val.Field(i).Interface())
		}
		return args
	}

	for i := 1; i < numField; i++ {
		args = append(args, val.Field(i).Interface())
	}
	args = append(args, val.Field(0).Interface())
	return args
}

func StructArgsList[T any](list []*T, idLast bool) [][]any {
	argsList := make([][]any, 0, len(list))
	for _, item := range list {
		argsList = append(argsList, StructArgs(item, idLast))
	}
	return argsList
}

func DeleteByID(d *DB, table string, id any) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE id=?", table)
	_, err := d.Exec(sql, []any{id}, true)
	return err
}

func DeleteWhere(d *DB, table string, where string, args ...any) error {
	sql := fmt.Sprintf("DELETE FROM %s", table)
	if where != "" {
		sql += " WHERE " + where
	}
	_, err := d.Exec(sql, args, true)
	return err
}

func Delete[T any](d *DB, table string, obj *T) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	id := val.Field(0).Interface()
	return DeleteByID(d, table, id)
}

func DeleteMany[T any](d *DB, table string, list []*T) error {
	if len(list) == 0 {
		return nil
	}

	ids := make([]any, 0, len(list))
	for _, obj := range list {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		ids = append(ids, val.Field(0).Interface())
	}

	placeholders := make([]string, len(ids))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	where := fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ", "))
	return DeleteWhere(d, table, where, ids...)
}

func SelectList[T any](d *DB, table string, query string, args ...any) ([]*T, error) {
	var zero T
	cols := StructCols(zero, false)

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), table)
	if query != "" {
		sql += " " + query
	}

	rows, err := d.Select(sql, args, true)
	if err != nil {
		return nil, err
	}

	list := make([]*T, 0, len(rows))
	for _, row := range rows {
		item := new(T)
		fillStruct(item, row)
		list = append(list, item)
	}
	return list, nil
}

func fillStruct(v any, row []any) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		if f.CanSet() {
			f.Set(reflect.ValueOf(row[i]).Convert(f.Type()))
		}
	}
}

func SelectBetween[T any](
	d *DB,
	table, field string,
	from, to any,
	order string,
) ([]*T, error) {

	query := fmt.Sprintf("WHERE %s BETWEEN ? AND ?", field)
	if order != "" {
		query += " ORDER BY " + order
	}
	return SelectList[T](d, table, query, from, to)
}

func SelectLast[T any](d *DB, table string) (*T, error) {
	var zero T
	cols := StructCols(zero, false)

	sql := fmt.Sprintf(
		"SELECT %s FROM %s ORDER BY id DESC LIMIT 1",
		strings.Join(cols, ", "),
		table,
	)

	rows, err := d.Select(sql, nil, true)
	if err != nil || len(rows) == 0 {
		return nil, err
	}

	item := new(T)
	fillStruct(item, rows[0])
	return item, nil
}

func SelectOne[T any](d *DB, table string) (*T, error) {
	var zero T
	cols := StructCols(zero, false)

	sql := fmt.Sprintf(
		"SELECT %s FROM %s LIMIT 1",
		strings.Join(cols, ", "),
		table,
	)

	rows, err := d.Select(sql, nil, true)
	if err != nil || len(rows) == 0 {
		return nil, err
	}

	item := new(T)
	fillStruct(item, rows[0])
	return item, nil
}

func SelectByID[T any](d *DB, table string, id int) (*T, error) {
	var zero T
	cols := StructCols(zero, false)

	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id=?",
		strings.Join(cols, ", "),
		table,
	)

	rows, err := d.Select(sql, []any{id}, true)
	if err != nil || len(rows) == 0 {
		return nil, err
	}

	item := new(T)
	fillStruct(item, rows[0])
	return item, nil
}

func Add[T any](d *DB, table string, v *T) (*T, error) {
	// struct 顺序（id 在 0），插入时跳过
	cols := StructCols(*v, false)[1:]
	args := StructArgs(*v, false)[1:]

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := d.Exec(sql, args, true)
	if err != nil {
		return nil, err
	}

	return SelectLast[T](d, table)
}

func AddMany[T any](d *DB, table string, list []*T) error {
	if len(list) == 0 {
		return nil
	}

	cols := StructCols(*list[0], false)[1:]

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	// 统一用 StructArgsList
	argsList := StructArgsList(list, false)
	for i := range argsList {
		argsList[i] = argsList[i][1:] // 去掉 id
	}

	return d.ExecMany(sql, argsList, true)
}

func Upsert[T any](d *DB, table string, obj *T) error {
	cols := StructCols(obj, true) // true 表示 id 放最后
	args := StructArgs(obj, true)

	// INSERT OR REPLACE INTO table (col1, col2, ..., id) VALUES (?, ?, ..., ?)
	sql := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(utils.MakePlaceholders(len(cols)), ", "),
	)

	_, err := d.Exec(sql, args, true)
	return err
}

func UpsertMany[T any](d *DB, table string, list []*T) error {
	if len(list) == 0 {
		return nil
	}

	cols := StructCols(list[0], true)      // 假设所有对象字段一致
	argsList := StructArgsList(list, true) // 批量参数

	sql := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(utils.MakePlaceholders(len(cols)), ", "),
	)

	return d.ExecMany(sql, argsList, true)
}
