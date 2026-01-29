package server

import (
	"diarygo/internal/bill"
	"diarygo/internal/db"
	"diarygo/internal/diary"
	"diarygo/internal/i18n"
	"diarygo/internal/interest"
	"diarygo/internal/note"
	"html/template"
	"net/http"
	"strconv"
)

var funcMap = template.FuncMap{
	"t": i18n.T,
}
var (
	loginTpl  *template.Template
	configTpl *template.Template
)

func RegisterRoutes() {
	DB := db.Get()
	configTpl = initTemplate("config.html", "web/templates/config.html", true)
	loginTpl = initTemplate("login.html", "web/templates/login.html", false)
	billRes := RegisterBillResource(DB)
	diaryRes := RegisterDiaryResource(DB)
	interestRes := RegisterInterestResource(DB)
	noteRes := RegisterNoteResource(DB)

	// -------------------- Web 页面 --------------------
	http.HandleFunc("/", loginPage)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/config", requireLogin(configPage))

	http.HandleFunc("/diary", PageHandler(diaryRes))
	http.HandleFunc("/bill", PageHandler(billRes))
	http.HandleFunc("/note", PageHandler(noteRes))
	http.HandleFunc("/interest", PageHandler(interestRes))

	// -------------------- REST API --------------------
	http.HandleFunc("/api/config/batch", requireLogin(configBatchAPI))
	http.HandleFunc("/api/config/change_password", requireLogin(configChangePasswordAPI))

	http.HandleFunc("/api/diary/list", ListHandler(diaryRes))
	http.HandleFunc("/api/diary/update", UpdateHandler(diaryRes))
	http.HandleFunc("/api/diary/export", ExportHandler(diaryRes))
	http.HandleFunc("/api/diary/import", ImportHandler(diaryRes))

	http.HandleFunc("/api/bill/list", ListHandler(billRes))
	http.HandleFunc("/api/bill/add", AddHandler(billRes))
	http.HandleFunc("/api/bill/update", UpdateByIDHandler(billRes))
	http.HandleFunc("/api/bill/delete", DeleteHandler(billRes))
	http.HandleFunc("/api/bill/export", ExportHandler(billRes))
	http.HandleFunc("/api/bill/import", ImportHandler(billRes))

	http.HandleFunc("/api/interest/list", ListHandler(interestRes))
	http.HandleFunc("/api/interest/add", AddHandler(interestRes))
	http.HandleFunc("/api/interest/update", UpdateByIDHandler(interestRes))
	http.HandleFunc("/api/interest/delete", DeleteHandler(interestRes))
	http.HandleFunc("/api/interest/export", ExportHandler(interestRes))
	http.HandleFunc("/api/interest/import", ImportHandler(interestRes))

	http.HandleFunc("/api/note/list", ListHandler(noteRes))
	http.HandleFunc("/api/note/add", AddHandler(noteRes))
	http.HandleFunc("/api/note/update", UpdateByIDHandler(noteRes))
	http.HandleFunc("/api/note/delete", DeleteHandler(noteRes))
	http.HandleFunc("/api/note/export", ExportHandler(noteRes))
	http.HandleFunc("/api/note/import", ImportHandler(noteRes))

	// 静态资源
	http.HandleFunc("/static/js/conf.js", confJsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
}

func RegisterDiaryResource(DB *db.DB) Resource[diary.Diary] {
	repo := diary.NewRepository(DB)
	return Resource[diary.Diary]{
		Name: diary.TABLE,
		Tpl:  initTemplate("diary.html", "web/templates/diary.html", true),
		Repo: repo,

		List: func(r *http.Request) (any, error) {
			year, _ := strconv.Atoi(r.URL.Query().Get("year"))
			month, _ := strconv.Atoi(r.URL.Query().Get("month"))
			return repo.GetMonthDiary(year, month)
		},

		Update: func(d *diary.Diary) error {
			return repo.Upsert(d)
		},
	}
}

func RegisterBillResource(DB *db.DB) Resource[bill.Bill] {
	repo := bill.NewRepository(DB)
	return Resource[bill.Bill]{
		Name: bill.TABLE,
		Tpl:  initTemplate("bill.html", "web/templates/bill.html", true),
		Repo: repo,

		List: func(r *http.Request) (any, error) {
			start, _ := strconv.Atoi(r.URL.Query().Get("start"))
			end, _ := strconv.Atoi(r.URL.Query().Get("end"))
			return repo.GetBetweenDates(start, end, "DESC")
		},
	}
}

func RegisterInterestResource(DB *db.DB) Resource[interest.Interest] {
	repo := interest.NewRepository(DB)
	return Resource[interest.Interest]{
		Name: interest.TABLE,
		Tpl:  initTemplate("interest.html", "web/templates/interest.html", true),
		Repo: repo,

		List: func(r *http.Request) (any, error) {
			sort, _ := strconv.Atoi(r.URL.Query().Get("sort"))
			return repo.GetBySort(sort)
		},
	}
}
func RegisterNoteResource(DB *db.DB) Resource[note.Note] {
	repo := note.NewRepository(DB)
	return Resource[note.Note]{
		Name: note.TABLE,
		Tpl:  initTemplate("note.html", "web/templates/note.html", true),
		Repo: repo,
	}
}
