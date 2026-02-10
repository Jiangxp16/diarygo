package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"diarygo/internal/config"
	"diarygo/internal/db"
)

const sessionCookie = "diarygo_session"
const sessionRedirect = "diarygo_redirect_after_login"

var rawKey = []byte("diarygo-session-key-32-bytes-OK!")
var rawSessionKey []byte
var aesSessionKey []byte

func initSessionKey() {
	cfgKey := config.GetRepository().Get("global", "session_key")
	if cfgKey != "" {
		rawSessionKey = []byte(cfgKey)
		fmt.Println("[session] use session_key from config")
	} else {
		rawSessionKey = make([]byte, 32)
		if _, err := rand.Read(rawSessionKey); err != nil {
			panic("failed to generate random session key")
		}
		fmt.Println("[session] no session_key in config, use random key (session will reset on restart)")
	}

	sum := sha256.Sum256(rawSessionKey)
	aesSessionKey = sum[:]
}

func encryptCookieValue(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesSessionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func decryptCookieValue(encoded string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesSessionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// -------------------- session helper --------------------
func checkSession(r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}

	password, err := decryptCookieValue(c.Value)
	if err != nil {
		return false
	}
	ok := config.GetRepository().CheckPassword(password)
	if ok {
		db.Key = password
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
	encry, err := encryptCookieValue(password)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   encry,
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
