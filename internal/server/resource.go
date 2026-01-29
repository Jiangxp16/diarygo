package server

import (
	"diarygo/internal/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

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

func render(w http.ResponseWriter, tpl *template.Template, data any) {
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
		render(w, res.Tpl, nil)
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
		jsonOK(w)
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
