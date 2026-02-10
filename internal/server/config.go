package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"diarygo/internal/config"
	"diarygo/internal/db"
	"diarygo/internal/i18n"
)

func confJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	cfg := config.GetRepository()

	config := map[string]any{
		"first_day_of_week": cfg.GetInt("global", "first_day_of_week", 1),
		"location":          cfg.Get("global", "location"),
		"language":          cfg.Get("global", "language"),
		//"login_expired":     cfg.GetInt("global", "login_expired", 600),
		"show_lunar":    cfg.GetBool("global", "show_lunar"),
		"show_bill":     cfg.GetBool("global", "show_bill"),
		"show_interest": cfg.GetBool("global", "show_interest"),
		"show_note":     cfg.GetBool("global", "show_note"),
		"show_sport":    cfg.GetBool("global", "show_sport"),
		"ui_default":    cfg.Get("global", "ui_default"),
		"logo":          cfg.Get("style", "logo"),
		"font":          cfg.Get("style", "font"),
	}
	jsonBytes, _ := json.Marshal(config)
	fmt.Fprintf(w, "window.APP_CONFIG = %s;", jsonBytes)
}

func configPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, configTpl, nil)
}

func configBatchAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Items []struct {
			Section string `json:"section"`
			Key     string `json:"key"`
			Value   string `json:"value"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "empty config items", http.StatusBadRequest)
		return
	}
	cfg := config.GetRepository()
	for _, item := range req.Items {
		if err := cfg.CheckValid(item.Section, item.Key, item.Value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	for _, item := range req.Items {
		if err := cfg.Set(item.Section, item.Key, item.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	for _, item := range req.Items {
		if item.Section == "global" && item.Key == "language" {
			i18n.Init(item.Value)
			break
		}
	}
	jsonOK(w)
}

func configChangePasswordAPI(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) > 16 {
		http.Error(w, "password too long", http.StatusBadRequest)
		return
	}

	cfg := config.GetRepository()

	d := db.Get()

	if err := cfg.ChangePassword(d, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w)
}
