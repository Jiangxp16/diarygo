package server

import (
	"diarygo/internal/config"
	"diarygo/internal/db"
	"diarygo/internal/entity/bill"
	"diarygo/internal/entity/diary"
	"diarygo/internal/entity/interest"
	"diarygo/internal/entity/note"
	"diarygo/internal/entity/sport"
	"diarygo/internal/i18n"
	"diarygo/internal/utils"

	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var funcMap = template.FuncMap{
	"t": i18n.T,
}

var (
	loginTpl  *template.Template
	configTpl *template.Template
)

func registerRoutes() {
	DB := db.Get()
	configTpl = initTemplate("config.html", "web/templates/config.html", true)
	loginTpl = initTemplate("login.html", "web/templates/login.html", false)
	billRes := RegisterBillResource(DB)
	diaryRes := RegisterDiaryResource(DB)
	interestRes := RegisterInterestResource(DB)
	noteRes := RegisterNoteResource(DB)
	sportRes := RegisterSportResource(DB)

	// -------------------- Web Page --------------------
	http.HandleFunc("/", loginPage)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/config", requireLogin(configPage))

	http.HandleFunc("/diary", PageHandler(diaryRes))
	http.HandleFunc("/bill", PageHandler(billRes))
	http.HandleFunc("/note", PageHandler(noteRes))
	http.HandleFunc("/interest", PageHandler(interestRes))
	http.HandleFunc("/sport", PageHandler(sportRes))

	// -------------------- REST API --------------------
	http.HandleFunc("/api/ping", requireLogin(PingHandler))
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

	http.HandleFunc("/api/sport/list", ListHandler(sportRes))
	http.HandleFunc("/api/sport/add", AddHandler(sportRes))
	http.HandleFunc("/api/sport/update", UpdateByIDHandler(sportRes))
	http.HandleFunc("/api/sport/delete", DeleteHandler(sportRes))
	http.HandleFunc("/api/sport/export", ExportHandler(sportRes))
	http.HandleFunc("/api/sport/import", ImportHandler(sportRes))

	http.HandleFunc("/static/js/conf.js", confJsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func jsonOK(w http.ResponseWriter) {
	jsonRes(w, map[string]any{"ok": true})
}

func jsonRes(w http.ResponseWriter, res any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["CurrentPath"] = r.URL.Path
	if err := tpl.ExecuteTemplate(w, "background.html", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func respSheetHead(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filename))
}

func initTemplate(htmlName string, htmlFile string, withBase bool) *template.Template {
	tpl := template.New(htmlName).Funcs(funcMap)
	if withBase {
		tpl = template.Must(tpl.ParseFiles(
			"web/base/background.html",
			"web/base/navbar.html",
			"web/base/i18n.html",
			htmlFile,
		))
	} else {
		tpl = template.Must(tpl.ParseFiles(htmlFile))
	}
	return tpl
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	jsonOK(w)
}

type Resource[T any] struct {
	Name       string
	Tpl        *template.Template
	Repo       any
	List       func(r *http.Request) (any, error)
	Add        func(obj *T) (*T, error)
	Update     func(obj *T) error
	UpdateByID func(id int, params map[string]any) error
	DeleteByID func(id int) error
	Export     func(w http.ResponseWriter) error
	Import     func(reader io.Reader) error
}

type CanList[T any] interface {
	List() ([]*T, error)
}

type CanAdd[T any] interface {
	Add(*T) (*T, error)
}

type CanUpdate[T any] interface {
	Update(*T) error
}

type CanUpdateByID interface {
	UpdateByID(id int, params map[string]any) error
}

type CanDelete interface {
	DeleteByID(int) error
}

type CanExport interface {
	Export(w io.Writer) error
}

type CanImport interface {
	Import(r io.Reader) error
}

func PageHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		render(w, r, res.Tpl, nil)
	})
}

func ListHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		if res.List != nil {
			data, err := res.List(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonRes(w, data)
			return
		}
		repo, ok := res.Repo.(CanList[T])
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		data, err := repo.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonRes(w, data)
	})
}

func AddHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		var obj T
		if err := decodeJSON(r, &obj); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if res.Add != nil {
			if _, err := res.Add(&obj); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanAdd[T])
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if _, err := repo.Add(&obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w)
	})
}

func UpdateHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		var obj T
		if err := decodeJSON(r, &obj); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if res.Update != nil {
			if err := res.Update(&obj); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanUpdate[T])
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if err := repo.Update(&obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w)
	})
}

func UpdateByIDHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := decodeJSON(r, &raw); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		rawID := raw["id"]
		delete(raw, "id")

		id, err := utils.NormalizeID(rawID)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if res.UpdateByID != nil {
			if err := res.UpdateByID(id, raw); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanUpdateByID)
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if err := repo.UpdateByID(id, raw); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w)
	})
}

func DeleteHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if res.DeleteByID != nil {
			if err := res.DeleteByID(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanDelete)
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if err := repo.DeleteByID(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w)
	})
}

func ExportHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := fmt.Sprintf("%s_%s.xlsx", res.Name, utils.GetNowYYMMDD())
		respSheetHead(w, filename)
		if res.Export != nil {
			if err := res.Export(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanExport)
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if err := repo.Export(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func ImportHandler[T any](res Resource[T]) http.HandlerFunc {
	return requireLogin(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		if res.Import != nil {
			if err := res.Import(file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w)
			return
		}
		repo, ok := res.Repo.(CanImport)
		if !ok {
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		if err := repo.Import(file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w)
	})
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
func RegisterSportResource(DB *db.DB) Resource[sport.Sport] {
	repo := sport.NewRepository(DB)
	return Resource[sport.Sport]{
		Name: sport.TABLE,
		Tpl:  initTemplate("sport.html", "web/templates/sport.html", true),
		Repo: repo,
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	loginTpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	password := r.FormValue("password")
	cfg := config.GetRepository()
	ok := cfg.CheckPassword(password)
	if !ok {
		stored := cfg.GetPassword()
		if stored == "" {
			cfg.SetPassword(password)
			ok = true
		}
	}
	if ok {
		setSession(w, password)

		if redirect := getRedirectAfterLogin(r); redirect != "" {
			clearRedirectAfterLogin(w)
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}

		uiDefault := cfg.Get("global", "ui_default")
		http.Redirect(w, r, "/"+uiDefault, http.StatusSeeOther)
		return
	}
	loginTpl.Execute(w, map[string]string{"Error": "Password incorrect"})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if shouldRecordRedirect(r) {
			setRedirectAfterLogin(w, r.URL.RequestURI())
		}
		if !checkSession(r) {
			if isAPIRequest(r) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			return
		}
		next(w, r)
	}
}
