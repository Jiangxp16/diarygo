package server

import (
	"net/http"
	"time"

	"diarygo/internal/config"
	"diarygo/internal/db"
)

const sessionCookie = "diarygo_session"

// -------------------- session helper --------------------
func checkSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	ok := config.GetRepository().CheckPassword(cookie.Value)
	if ok {
		db.Key = cookie.Value
	}
	return ok
}

func setSession(w http.ResponseWriter, password string) {
	sessionDuration, _ := time.ParseDuration(
		config.GetRepository().Get("global", "login_expired"),
	)
	if sessionDuration == 0 {
		sessionDuration = 600 * time.Second
	}
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   password,
		Expires: time.Now().Add(sessionDuration),
		Path:    "/",
	})
}

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	})
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

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkSession(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
