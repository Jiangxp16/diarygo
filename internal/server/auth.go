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
	if err != nil {
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
