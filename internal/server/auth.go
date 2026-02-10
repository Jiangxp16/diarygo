package server

import (
	"net/http"
	"strings"
	"time"

	"diarygo/internal/config"
	"diarygo/internal/db"
)

const sessionCookie = "diarygo_session"
const sessionRedirect = "diarygo_redirect_after_login"

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

func shouldRecordRedirect(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		return false
	}
	if r.URL.Path == "/" {
		return false
	}
	return true
}

func setRedirectAfterLogin(w http.ResponseWriter, url string) {
	http.SetCookie(w, &http.Cookie{
		Name:  sessionRedirect,
		Value: url,
		Path:  "/",
	})
}

func getRedirectAfterLogin(r *http.Request) string {
	c, err := r.Cookie(sessionRedirect)
	if err != nil {
		return ""
	}
	return c.Value
}

func clearRedirectAfterLogin(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionRedirect,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
